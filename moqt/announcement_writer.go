package moqt

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSendAnnounceStream(stream quic.Stream, prefix string) *AnnouncementWriter {
	sas := &AnnouncementWriter{
		prefix:    prefix,
		stream:    stream,
		streamCtx: stream.Context(),
		actives:   make(map[string]*Announcement),
		initCh:    make(chan struct{}, 1), // Buffered to avoid blocking on init
	}

	return sas
}

type AnnouncementWriter struct {
	mu sync.RWMutex

	prefix    string
	stream    quic.Stream
	streamCtx context.Context

	actives map[string]*Announcement

	initOnce sync.Once
	initCh   chan struct{}
}

func (sas *AnnouncementWriter) init(init []*Announcement) error {
	var err error
	sas.initOnce.Do(func() {
		sas.mu.Lock()
		defer sas.mu.Unlock()

		if sas.streamCtx.Err() != nil {
			reason := context.Cause(sas.streamCtx)
			var strErr *quic.StreamError
			if errors.As(reason, &strErr) {
				err = &AnnounceError{
					StreamError: strErr,
				}
				return
			}
			err = reason
			return
		}

		suffixes := make([]string, 0, len(init))
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
					return // Already active, no need to re-announce
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
				if current, ok := sas.actives[suffix]; ok && current == new {
					delete(sas.actives, suffix)
				}

				if sas.stream.Context().Err() != nil {
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
			slog.Error("failed to send ANNOUNCE_INIT message", "error", err)
			var strErr *quic.StreamError
			if errors.As(err, &strErr) {
				err = &AnnounceError{
					StreamError: strErr,
				}
			}

			return
		}

		initCh := sas.initCh
		close(initCh)
	})

	return err
}

func (sas *AnnouncementWriter) SendAnnouncement(new *Announcement) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.streamCtx.Err() != nil {
		reason := context.Cause(sas.streamCtx)
		var strErr *quic.StreamError
		if errors.As(reason, &strErr) {
			return &AnnounceError{
				StreamError: strErr,
			}
		}
		return reason
	}

	if sas.initCh != nil {
		<-sas.initCh
		sas.initCh = nil
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

	sas.stream.CancelRead(quic.StreamErrorCode(InternalAnnounceErrorCode)) // TODO: Use a specific error code if needed
	return sas.stream.Close()
}

func (sas *AnnouncementWriter) CloseWithError(code AnnounceErrorCode) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	strErrCode := quic.StreamErrorCode(code)
	sas.stream.CancelWrite(strErrCode)
	sas.stream.CancelRead(strErrCode)

	return nil
}
