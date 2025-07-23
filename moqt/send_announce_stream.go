package moqt

import (
	"context"
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSendAnnounceStream(stream quic.Stream, prefix string) *AnnouncementWriter {
	sas := &AnnouncementWriter{
		prefix:  prefix,
		stream:  stream,
		actives: make(map[string]*Announcement),
	}

	return sas
}

type AnnouncementWriter struct {
	mu sync.RWMutex

	prefix string
	stream quic.Stream

	actives map[string]*Announcement
}

func (sas *AnnouncementWriter) SendAnnouncement(new *Announcement) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.stream.Context().Err() != nil {
		reason := context.Cause(sas.stream.Context())
		var strErr *quic.StreamError
		if errors.As(reason, &strErr) {
			return &AnnounceError{
				StreamError: strErr,
			}
		}
		return reason
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
		if old != new {
			old.End()
		}
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
	go func() {
		<-new.AwaitEnd()

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
	}()

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
