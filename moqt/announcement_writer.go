package moqt

import (
	"context"
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

func newAnnouncementWriter(stream quic.Stream, prefix prefix) *AnnouncementWriter {
	if !isValidPrefix(prefix) {
		panic("invalid prefix for AnnouncementWriter")
	}

	sas := &AnnouncementWriter{
		prefix:   prefix,
		stream:   stream,
		ctx:      context.WithValue(stream.Context(), &biStreamTypeCtxKey, message.StreamTypeAnnounce),
		actives:  make(map[suffix]*Announcement),
		endFuncs: make(map[suffix]func()),
		initCh:   make(chan struct{}),
	}

	return sas
}

type AnnouncementWriter struct {
	prefix prefix
	stream quic.Stream
	ctx    context.Context

	mu       sync.RWMutex
	actives  map[suffix]*Announcement
	endFuncs map[suffix]func()

	initCh   chan struct{}
	initOnce sync.Once
}

func (aw *AnnouncementWriter) init(init map[*Announcement]struct{}) error {
	var err error
	aw.initOnce.Do(func() {
		if aw.ctx.Err() != nil {
			err = Cause(aw.ctx)
			return
		}

		// We'll collect suffixes from the staged map to avoid duplicates
		// and reduce temporary allocations.
		var suffixes []suffix
		oldAnns := make([]*Announcement, 0) // Store old announcements to end later

		// We'll batch updates to aw.actives and aw.endFuncs to reduce lock churn.
		// Pre-size the map based on input size to reduce rehashing.
		toSet := make(map[suffix]struct {
			ann *Announcement
			end func()
		}, len(init))

		// Cache frequently used fields locally to avoid repeated struct field loads
		ctx := aw.ctx
		stream := aw.stream

		for new := range init {
			if !new.IsActive() {
				continue // Skip non-active announcements
			}

			suffix, ok := new.BroadcastPath().GetSuffix(aw.prefix)
			if !ok {
				continue // Invalid path, skip
			}

			// If we've already staged a new announcement for this suffix,
			// treat the previously staged one as old and replace it.
			if prev, ok := toSet[suffix]; ok {
				// staged previous announcement should be ended
				oldAnns = append(oldAnns, prev.ann)
			} else if old, ok := aw.actives[suffix]; ok {
				// Otherwise, if there's an existing active announcement on the writer,
				// mark it as old so it will be ended and replaced.
				if old != new {
					oldAnns = append(oldAnns, old)
				} else {
					// already active (same instance) - skip
					continue
				}
			}

			// capture loop variables to avoid closure capture bugs
			n := new
			s := suffix
			endFunc := func() {
				aw.mu.Lock()
				if cur, ok := aw.actives[s]; ok && cur == n {
					delete(aw.actives, s)
					delete(aw.endFuncs, s)
				}
				aw.mu.Unlock()

				// Send ENDED message without acquiring any locks to avoid deadlock
				if ctx.Err() != nil {
					return
				}

				// Send the message directly using cached stream
				message.AnnounceMessage{
					AnnounceStatus: message.ENDED,
					TrackSuffix:    s,
				}.Encode(stream)
			}

			// Stage for batched set (overwrites any previous staged announcement)
			toSet[suffix] = struct {
				ann *Announcement
				end func()
			}{ann: new, end: endFunc}
		}

		// After staging, build the unique suffix list from toSet
		if len(toSet) > 0 {
			suffixes = make([]suffix, 0, len(toSet))
			for s := range toSet {
				suffixes = append(suffixes, s)
			}
		}

		// Apply batched updates under a single lock
		if len(toSet) > 0 {
			aw.mu.Lock()
			for s, v := range toSet {
				aw.actives[s] = v.ann
				aw.endFuncs[s] = v.end
			}
			aw.mu.Unlock()

			// Register OnEnd handlers after we've set the maps so end handlers see maps populated
			for _, v := range toSet {
				v.ann.OnEnd(v.end)
			}
		}

		// End old announcements after releasing the mutex
		for _, old := range oldAnns {
			// end() triggers OnEnd callbacks to send ENDED messages, etc.
			old.end()
		}

		if ctx.Err() != nil {
			err = Cause(ctx)
			return
		}

		err = message.AnnounceInitMessage{
			Suffixes: suffixes,
		}.Encode(stream)
		if err != nil {
			var strErr *quic.StreamError
			if errors.As(err, &strErr) {
				err = &AnnounceError{
					StreamError: strErr,
				}
			}

			return
		}

		aw.mu.Lock()
		if aw.initCh != nil {
			close(aw.initCh)
			aw.initCh = nil
		}
		aw.mu.Unlock()
	})

	return err
}

func (aw *AnnouncementWriter) SendAnnouncement(new *Announcement) error {
	// Cache frequently used fields
	ctx := aw.ctx
	stream := aw.stream

	if ctx.Err() != nil {
		return Cause(ctx)
	}

	// Wait for initialization outside of lock
	aw.mu.RLock()
	initCh := aw.initCh
	aw.mu.RUnlock()

	if initCh != nil {
		<-initCh
	}

	if ctx.Err() != nil {
		return Cause(ctx)
	}

	// Get suffix for this announcement
	suffix, ok := new.BroadcastPath().GetSuffix(aw.prefix)
	if !ok {
		return errors.New("invalid broadcast path")
	}

	if !new.IsActive() {
		return nil // No need to send inactive announcements
	}

	// Check for previous announcement and get old endFunc
	aw.mu.Lock()
	old, exists := aw.actives[suffix]
	if exists && old == new {
		aw.mu.Unlock()
		return nil // Already active, no need to re-announce
	}

	var oldEndFunc func()
	if exists {
		oldEndFunc = aw.endFuncs[suffix]
	}
	aw.mu.Unlock()

	// End old announcement outside of lock
	if oldEndFunc != nil {
		old.end() //
	}

	if aw.ctx.Err() != nil {
		return Cause(aw.ctx)
	}

	// Create reusable AnnounceMessage
	am := message.AnnounceMessage{
		AnnounceStatus: message.ACTIVE,
		TrackSuffix:    suffix,
	}

	// Encode and send ACTIVE announcement using cached stream
	err := am.Encode(stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			return &AnnounceError{
				StreamError: strErr,
			}
		}

		return err
	}

	endFunc := func() {
		aw.mu.Lock()
		if current, exists := aw.actives[suffix]; exists && current == new {
			delete(aw.actives, suffix)
			delete(aw.endFuncs, suffix)
		}
		aw.mu.Unlock()

		// Send ENDED message without holding locks
		if ctx.Err() != nil {
			return
		}

		message.AnnounceMessage{
			AnnounceStatus: message.ENDED,
			TrackSuffix:    suffix,
		}.Encode(stream)
	}

	// Update actives map atomically
	aw.mu.Lock()
	aw.actives[suffix] = new
	aw.endFuncs[suffix] = endFunc
	aw.mu.Unlock()

	// Watch for announcement end in background
	new.OnEnd(endFunc)

	return nil
}

func (aw *AnnouncementWriter) Close() error {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	aw.actives = nil
	aw.endFuncs = nil

	if aw.initCh != nil {
		close(aw.initCh)
		aw.initCh = nil
	}

	aw.stream.CancelRead(quic.StreamErrorCode(InternalAnnounceErrorCode)) // TODO: Use a specific error code if needed
	return aw.stream.Close()
}

func (aw *AnnouncementWriter) CloseWithError(code AnnounceErrorCode) error {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	aw.actives = nil
	aw.endFuncs = nil

	if aw.initCh != nil {
		close(aw.initCh)
		aw.initCh = nil
	}

	strErrCode := quic.StreamErrorCode(code)
	aw.stream.CancelWrite(strErrCode)
	aw.stream.CancelRead(strErrCode)

	return nil
}

func (aw *AnnouncementWriter) Context() context.Context {
	return aw.ctx
}
