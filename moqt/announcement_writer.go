package moqt

import (
	"context"
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newAnnouncementWriter(stream quic.Stream, prefix prefix) *AnnouncementWriter {
	if !isValidPrefix(prefix) {
		panic("invalid prefix for AnnouncementWriter")
	}

	sas := &AnnouncementWriter{
		prefix:  prefix,
		stream:  stream,
		ctx:     context.WithValue(stream.Context(), &biStreamTypeCtxKey, message.StreamTypeAnnounce),
		actives: make(map[suffix]*Announcement),
		initCh:  make(chan struct{}, 1), // Buffered to avoid blocking on init
		cleanCh: make(chan struct{}, 1), // Buffered to avoid blocking on clean
	}

	go func() {
		for range sas.cleanCh {
			if sas.Context().Err() != nil {
				return // Context cancelled, exit goroutine
			}

			sas.mu.Lock()
			// Remove inactive announcements from actives map
			for suffix, ann := range sas.actives {
				if !ann.IsActive() {
					delete(sas.actives, suffix)
				}
			}
			sas.mu.Unlock()
		}
	}()

	return sas
}

type AnnouncementWriter struct {
	mu sync.RWMutex

	prefix prefix
	stream quic.Stream
	ctx    context.Context

	actives map[suffix]*Announcement

	initCh   chan struct{}
	initOnce sync.Once

	cleanCh chan struct{}
}

func (sas *AnnouncementWriter) init(init []*Announcement) error {
	var err error
	sas.initOnce.Do(func() {
		sas.mu.Lock()
		defer sas.mu.Unlock()

		if sas.ctx.Err() != nil {
			// sas.mu.Unlock()
			err = Cause(sas.ctx)
			return
		}

		suffixes := make([]suffix, 0, len(init))
		oldAnnouncements := make([]*Announcement, 0) // Store old announcements to end later

		for _, new := range init {
			if !new.IsActive() {
				continue // Skip non-active announcements
			}

			suffix, ok := new.BroadcastPath().GetSuffix(sas.prefix)
			if !ok {
				continue // Invalid path, skip
			}

			// Check for previous announcement
			if old, ok := sas.actives[suffix]; ok {
				if old == new {
					continue // Already active, no need to re-announce
				}
				oldAnnouncements = append(oldAnnouncements, old)
			}

			sas.actives[suffix] = new
			suffixes = append(suffixes, suffix)

			// Watch for announcement end in background
			new.OnEnd(func() {
				// Signal cleanup goroutine to handle map cleanup
				select {
				case sas.cleanCh <- struct{}{}:
				default:
				}

				// Send ENDED message without acquiring any locks to avoid deadlock
				if sas.Context().Err() != nil {
					return
				}

				// Send the message directly
				message.AnnounceMessage{
					AnnounceStatus: message.ENDED,
					TrackSuffix:    suffix,
				}.Encode(sas.stream)
			})
		}

		// Unlock before calling End() on old announcements to avoid deadlock
		sas.mu.Unlock()

		// End old announcements after releasing the mutex
		for _, old := range oldAnnouncements {
			old.End()
		}

		sas.mu.Lock()

		if sas.ctx.Err() != nil {
			err = Cause(sas.ctx)
			return
		}

		err = message.AnnounceInitMessage{
			Suffixes: suffixes,
		}.Encode(sas.stream)
		if err != nil {
			var strErr *quic.StreamError
			if errors.As(err, &strErr) {
				err = &AnnounceError{
					StreamError: strErr,
				}
			}

			return
		}

		close(sas.initCh)
		sas.initCh = nil
	})

	return err
}

func (sas *AnnouncementWriter) SendAnnouncement(new *Announcement) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.ctx.Err() != nil {
		return Cause(sas.ctx)
	}

	if sas.initCh != nil {
		initCh := sas.initCh
		sas.mu.Unlock()
		<-initCh
		sas.mu.Lock()
	}

	if sas.ctx.Err() != nil {
		return Cause(sas.ctx)
	}

	if !new.IsActive() {
		return errors.New("announcement must be active")
	}

	// Get suffix for this announcement
	suffix, ok := new.BroadcastPath().GetSuffix(sas.prefix)
	if !ok {
		return errors.New("invalid broadcast path")
	}

	var oldAnnouncement *Announcement
	// Check for previous announcement
	if old, ok := sas.actives[suffix]; ok {
		if old == new {
			return nil // Already active, no need to re-announce
		}
		oldAnnouncement = old
	}

	sas.actives[suffix] = new

	// Unlock before calling End() on old announcement to avoid deadlock
	sas.mu.Unlock()

	// End old announcement after releasing the mutex
	if oldAnnouncement != nil {
		oldAnnouncement.End()
	}

	// Re-acquire lock for the remaining operations
	sas.mu.Lock()

	if sas.ctx.Err() != nil {
		return Cause(sas.ctx)
	}

	// Create reusable AnnounceMessage
	am := message.AnnounceMessage{
		AnnounceStatus: message.ACTIVE,
		TrackSuffix:    suffix,
	}

	// Encode and send ACTIVE announcement
	err := am.Encode(sas.stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			return &AnnounceError{
				StreamError: strErr,
			}
		}

		return err
	}

	// Watch for announcement end in background
	new.OnEnd(func() {
		// Signal cleanup goroutine to handle map cleanup
		select {
		case sas.cleanCh <- struct{}{}:
		default:
		}

		// Send ENDED message without acquiring any locks to avoid deadlock
		// We'll check context separately without holding locks
		if sas.stream.Context().Err() != nil {
			return
		}

		// Send the message directly - any errors will be handled by the stream
		message.AnnounceMessage{
			AnnounceStatus: message.ENDED,
			TrackSuffix:    suffix,
		}.Encode(sas.stream)
	})

	return nil
}

func (sas *AnnouncementWriter) Close() error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	sas.actives = nil

	if sas.initCh != nil {
		close(sas.initCh)
		sas.initCh = nil
	}

	// Close cleanup channel to stop cleanup goroutine
	if sas.cleanCh != nil {
		close(sas.cleanCh)
		sas.cleanCh = nil
	}

	sas.stream.CancelRead(quic.StreamErrorCode(InternalAnnounceErrorCode)) // TODO: Use a specific error code if needed
	return sas.stream.Close()
}

func (sas *AnnouncementWriter) CloseWithError(code AnnounceErrorCode) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	sas.actives = nil

	if sas.initCh != nil {
		close(sas.initCh)
		sas.initCh = nil
	}

	// Close cleanup channel to stop cleanup goroutine
	if sas.cleanCh != nil {
		close(sas.cleanCh)
		sas.cleanCh = nil
	}

	strErrCode := quic.StreamErrorCode(code)
	sas.stream.CancelWrite(strErrCode)
	sas.stream.CancelRead(strErrCode)

	return nil
}

func (sas *AnnouncementWriter) Context() context.Context {
	return sas.ctx
}
