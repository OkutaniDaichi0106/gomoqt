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
	}

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
}

func (sas *AnnouncementWriter) init(init []*Announcement) error {
	var err error
	sas.initOnce.Do(func() {
		sas.mu.Lock()
		defer sas.mu.Unlock()

		if sas.ctx.Err() != nil {
			err = Cause(sas.ctx)
			return
		}

		suffixes := make([]suffix, 0, len(init))
		for _, new := range init {
			if !new.IsActive() {
				continue // Skip non-active announcements
			}

			suffix, ok := new.BroadcastPath().GetSuffix(sas.prefix)
			if !ok {
				continue // Invalid path, skip
			}

			// Cancel previous announcement if exists
			if old, ok := sas.actives[suffix]; ok {
				if old == new {
					continue // Already active, no need to re-announce
				}
				old.End()
			}

			sas.actives[suffix] = new

			suffixes = append(suffixes, suffix)

			// Watch for announcement end in background
			new.OnEnd(func() {
				sas.mu.Lock()
				defer sas.mu.Unlock()

				// Remove from actives only if it's still the same announcement
				if sas.actives == nil {
					return // Already closed
				}

				if current, ok := sas.actives[suffix]; ok && current == new {
					delete(sas.actives, suffix)
				}

				if sas.ctx.Err() != nil {
					return
				}

				// Encode and send ENDED announcement
				err := message.AnnounceMessage{
					AnnounceStatus: message.ENDED,
					TrackSuffix:    suffix,
				}.Encode(sas.stream)
				if err != nil {
					return
				}
			})
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
		<-sas.initCh
	}

	if !new.IsActive() {
		return errors.New("announcement must be active")
	}

	// Get suffix for this announcement
	suffix, ok := new.BroadcastPath().GetSuffix(sas.prefix)
	if !ok {
		return errors.New("invalid broadcast path")
	}

	// Cancel previous announcement if exists
	if old, ok := sas.actives[suffix]; ok {
		if old == new {
			return nil // Already active, no need to re-announce
		}
		old.End()
	}

	sas.actives[suffix] = new

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
		sas.mu.Lock()
		defer sas.mu.Unlock()

		// Remove from actives only if it's still the same announcement
		if current, ok := sas.actives[suffix]; ok && current == new {
			delete(sas.actives, suffix)
		}

		if sas.stream.Context().Err() != nil {
			return
		}

		// Reuse the same AnnounceMessage, just change status
		am.AnnounceStatus = message.ENDED

		// Encode and send ENDED announcement
		err := am.Encode(sas.stream)
		if err != nil {
			return
		}
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

	strErrCode := quic.StreamErrorCode(code)
	sas.stream.CancelWrite(strErrCode)
	sas.stream.CancelRead(strErrCode)

	return nil
}

func (sas *AnnouncementWriter) Context() context.Context {
	return sas.ctx
}
