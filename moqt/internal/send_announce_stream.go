package internal

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newSendAnnounceStream(stream transport.Stream, apm message.AnnouncePleaseMessage) *SendAnnounceStream {
	return &SendAnnounceStream{
		AnnouncePleaseMessage: apm,
		Stream:                stream,
	}
}

type SendAnnounceStream struct {
	AnnouncePleaseMessage message.AnnouncePleaseMessage

	Stream transport.Stream
	mu     sync.RWMutex

	announcements []message.AnnounceMessage

	closed   bool
	closeErr error
}

func (sas *SendAnnounceStream) SetActiveAnnouncement(suffix string) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	slog.Debug("sending announcements", slog.Any("trackPathSuffix", suffix))

	am := message.AnnounceMessage{
		AnnounceStatus: message.ACTIVE,
		TrackSuffix:    suffix,
	}

	sas.announcements = append(sas.announcements, am)

	slog.Debug("sent announcements", slog.Any("announcements", am))

	return nil
}

func (sas *SendAnnounceStream) SetEndedAnnouncement(suffix string) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	slog.Debug("sending announcements", slog.Any("trackPathSuffix", suffix))

	am := message.AnnounceMessage{
		AnnounceStatus: message.ENDED,
		TrackSuffix:    suffix,
	}

	sas.announcements = append(sas.announcements, am)

	return nil
}

func (sas *SendAnnounceStream) SendAnnouncements() error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	slog.Debug("sending announcements")

	// Calculate the total length of the ANNOUNCE messages
	var len int
	for _, am := range sas.announcements {
		len += am.Len()
	}
	am := message.AnnounceMessage{
		AnnounceStatus: message.LIVE,
	}
	len += am.Len()

	// Create a buffer
	buf := bytes.NewBuffer(make([]byte, 0, len))

	// Encode the ANNOUNCE messages
	for _, am := range sas.announcements {
		// Encode the ANNOUNCE message
		_, err := am.Encode(buf)
		if err != nil {
			slog.Error("failed to send an ANNOUNCE message", slog.String("error", err.Error()))
			return err
		}
	}

	// Send the ANNOUNCE messages
	_, err := sas.Stream.Write(buf.Bytes())
	if err != nil {
		slog.Error("failed to send an ANNOUNCE message", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (sas *SendAnnounceStream) Close() error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	return sas.close()
}

func (sas *SendAnnounceStream) CloseWithError(err error) error { // TODO
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if err == nil {
		return sas.close()
	}

	var annerr AnnounceError
	if !errors.As(err, &annerr) {
		annerr = ErrInternalError
	}

	code := transport.StreamErrorCode(annerr.AnnounceErrorCode())
	sas.Stream.CancelRead(code)
	sas.Stream.CancelWrite(code)

	slog.Debug("closed a send announce stream with an error", slog.String("error", err.Error()))

	return nil
}

func (sas *SendAnnounceStream) close() error {
	if sas.closed {
		if sas.closeErr != nil {
			return fmt.Errorf("stream has already closed due to: %w", sas.closeErr)
		}
		return errors.New("stream has already closed")
	}

	sas.closed = true
	return sas.Stream.Close()
}
