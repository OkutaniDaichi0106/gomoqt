package internal

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
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

	slog.Debug("sending active announcement", slog.String("trackPathSuffix", suffix))
	am := message.AnnounceMessage{
		AnnounceStatus: message.ACTIVE,
		TrackSuffix:    suffix,
	}
	sas.announcements = append(sas.announcements, am)
	slog.Debug("active announcement added", slog.Any("announcement", am), slog.Int("totalAnnouncements", len(sas.announcements)))
	return nil
}

func (sas *SendAnnounceStream) SetEndedAnnouncement(suffix string) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	slog.Debug("sending ended announcement", slog.String("trackPathSuffix", suffix))
	am := message.AnnounceMessage{
		AnnounceStatus: message.ENDED,
		TrackSuffix:    suffix,
	}
	sas.announcements = append(sas.announcements, am)
	slog.Debug("ended announcement added", slog.Any("announcement", am), slog.Int("totalAnnouncements", len(sas.announcements)))
	return nil
}

func (sas *SendAnnounceStream) SendAnnouncements() error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	slog.Debug("preparing to send announcements", slog.Int("announcementCount", len(sas.announcements)))

	// Calculate the total length of the ANNOUNCE messages
	var totalLen int
	for _, am := range sas.announcements {
		totalLen += am.Len()
	}
	am := message.AnnounceMessage{
		AnnounceStatus: message.LIVE,
	}
	totalLen += am.Len()

	// Create a buffer
	buf := bytes.NewBuffer(make([]byte, 0, totalLen))

	// Encode the ANNOUNCE messages
	for _, am := range sas.announcements {
		// Encode the ANNOUNCE message
		_, err := am.Encode(buf)
		if err != nil {
			slog.Error("failed to encode an ANNOUNCE message", slog.String("error", err.Error()))
			return err
		}
	}
	slog.Debug("encoded announcements", slog.Int("bufferLength", buf.Len()))

	// Send the ANNOUNCE messages
	_, err := sas.Stream.Write(buf.Bytes())
	if err != nil {
		slog.Error("failed to send ANNOUNCE messages", slog.String("error", err.Error()))
		return err
	}
	slog.Debug("sent announcements successfully", slog.Int("bytesSent", buf.Len()))

	return nil
}

func (sas *SendAnnounceStream) Close() error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return fmt.Errorf("stream already closed due to: %w", sas.closeErr)
		}
		return errors.New("stream already closed")
	}

	slog.Debug("closing stream")
	sas.closed = true
	err := sas.Stream.Close()
	if err != nil {
		slog.Error("error on stream close", slog.String("error", err.Error()))
	} else {
		slog.Debug("stream closed successfully")
	}
	return err
}

func (sas *SendAnnounceStream) CloseWithError(err error) error { // TODO
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if err == nil {
		err = ErrInternalError
	}

	var annerr AnnounceError
	if !errors.As(err, &annerr) {
		annerr = ErrInternalError
	}

	code := transport.StreamErrorCode(annerr.AnnounceErrorCode())
	sas.Stream.CancelRead(code)
	sas.Stream.CancelWrite(code)

	slog.Debug("closed send announce stream with error",
		slog.String("error", err.Error()),
		slog.String("cancel code", strconv.Itoa(int(code))))
	return nil
}
