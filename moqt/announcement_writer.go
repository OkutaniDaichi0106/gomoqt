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

		suffixes := make([]suffix, 0, len(init))
		oldAnns := make([]*Announcement, 0) // Store old announcements to end later

		for new := range init {
			if !new.IsActive() {
				continue // Skip non-active announcements
			}

			suffix, ok := new.BroadcastPath().GetSuffix(aw.prefix)
			if !ok {
				continue // Invalid path, skip
			}

			// Check for previous announcement
			if old, ok := aw.actives[suffix]; ok {
				if old == new {
					continue // Already active, no need to re-announce
				}
				// Always treat the existing one as the old one and replace with the new one
				oldAnns = append(oldAnns, old)
			}

			newEndFunc := func() {
				aw.mu.Lock()
				if cur, ok := aw.actives[suffix]; ok && cur == new {
					delete(aw.actives, suffix)
					delete(aw.endFuncs, suffix)
				}
				aw.mu.Unlock()

				// Send ENDED message without acquiring any locks to avoid deadlock
				if aw.Context().Err() != nil {
					return
				}

				// Send the message directly
				message.AnnounceMessage{
					AnnounceStatus: message.ENDED,
					TrackSuffix:    suffix,
				}.Encode(aw.stream)
			}

			aw.mu.Lock()
			aw.actives[suffix] = new
			aw.endFuncs[suffix] = newEndFunc
			aw.mu.Unlock()

			suffixes = append(suffixes, suffix)

			// Watch for announcement end in background
			new.OnEnd(newEndFunc)
		}

		// End old announcements after releasing the mutex
		for _, old := range oldAnns {
			// end() triggers OnEnd callbacks to send ENDED messages, etc.
			old.end()
		}

		if aw.ctx.Err() != nil {
			err = Cause(aw.ctx)
			return
		}

		err = message.AnnounceInitMessage{
			Suffixes: suffixes,
		}.Encode(aw.stream)
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
	if aw.ctx.Err() != nil {
		return Cause(aw.ctx)
	}

	// Wait for initialization outside of lock
	aw.mu.RLock()
	initCh := aw.initCh
	aw.mu.RUnlock()

	if initCh != nil {
		<-initCh
	}

	if aw.ctx.Err() != nil {
		return Cause(aw.ctx)
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
		old.end() // 旧アナウンスは終了させる（OnEnd 経由でクリーンアップ）
	}

	if aw.ctx.Err() != nil {
		return Cause(aw.ctx)
	}

	// Create reusable AnnounceMessage
	am := message.AnnounceMessage{
		AnnounceStatus: message.ACTIVE,
		TrackSuffix:    suffix,
	}

	// Encode and send ACTIVE announcement
	err := am.Encode(aw.stream)
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
		if aw.Context().Err() != nil {
			return
		}

		message.AnnounceMessage{
			AnnounceStatus: message.ENDED,
			TrackSuffix:    suffix,
		}.Encode(aw.stream)
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
