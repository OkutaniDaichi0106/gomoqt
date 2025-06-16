package moqt

import (
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type AnnouncementWriter interface {
	SendAnnouncement(announcement *Announcement) error
	Close() error
	CloseWithError(code AnnounceErrorCode) error
}

func newSendAnnounceStream(stream quic.Stream, prefix string) *sendAnnounceStream {
	sas := &sendAnnounceStream{
		prefix:  prefix,
		stream:  stream,
		actives: make(map[string]*Announcement),
	}

	return sas
}

var _ AnnouncementWriter = (*sendAnnounceStream)(nil)

type sendAnnounceStream struct {
	mu sync.RWMutex

	prefix string
	stream quic.Stream

	actives map[string]*Announcement

	closed   bool
	closeErr error
}

func (sas *sendAnnounceStream) SendAnnouncement(new *Announcement) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return sas.closeErr
		}
		return errors.New("stream already closed")
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
	_, err := am.Encode(sas.stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			sas.closed = true
			sas.closeErr = &AnnounceError{
				StreamError: strErr,
			}

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

		if sas.closed {
			return
		}

		// Reuse the same AnnounceMessage, just change status
		am.AnnounceStatus = message.ENDED

		// Encode and send ENDED announcement
		_, err := am.Encode(sas.stream)
		if err != nil {
			var strErr *quic.StreamError
			if errors.As(err, &strErr) {
				sas.closed = true
				sas.closeErr = &AnnounceError{
					StreamError: strErr,
				}
			}
		}
	}()

	return nil
}

func (sas *sendAnnounceStream) Close() error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		return sas.closeErr
	}

	sas.closed = true

	err := sas.stream.Close()
	if err != nil {
		return err
	}

	return nil
}

func (sas *sendAnnounceStream) CloseWithError(code AnnounceErrorCode) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		return sas.closeErr
	}

	sas.closed = true

	strErrCode := quic.StreamErrorCode(code)
	sas.closeErr = &AnnounceError{
		StreamError: &quic.StreamError{
			StreamID:  sas.stream.StreamID(),
			ErrorCode: strErrCode,
		},
	}

	sas.stream.CancelWrite(strErrCode)
	sas.stream.CancelRead(strErrCode)

	return nil
}
