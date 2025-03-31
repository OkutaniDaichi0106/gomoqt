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
		announced:             make(map[string]message.AnnounceMessage),
	}
}

type SendAnnounceStream struct {
	AnnouncePleaseMessage message.AnnouncePleaseMessage

	Stream transport.Stream
	mu     sync.RWMutex

	pending map[string]message.AnnounceMessage

	announced map[string]message.AnnounceMessage

	closed   bool
	closeErr error
}

func (sas *SendAnnounceStream) SetActiveAnnouncement(suffix string) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return fmt.Errorf("stream already closed due to: %w", sas.closeErr)
		}
		return errors.New("stream already closed")
	}

	if am, ok := sas.announced[suffix]; ok {
		if am.AnnounceStatus == message.ACTIVE {
			return fmt.Errorf("duplicated announcement: %s", suffix)
		}
	}

	//
	if am, ok := sas.pending[suffix]; ok {
		if am.AnnounceStatus == message.ACTIVE {
			// Skip
			slog.Warn("active announcement already set", "track_suffix", suffix)
			return nil
		}
	}

	am := message.AnnounceMessage{
		AnnounceStatus: message.ACTIVE,
		TrackSuffix:    suffix,
	}

	sas.pending[suffix] = am

	slog.Debug("set active announcement", slog.String("track_suffix", suffix))

	return nil
}

func (sas *SendAnnounceStream) SetEndedAnnouncement(suffix string) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return fmt.Errorf("stream already closed due to: %w", sas.closeErr)
		}
		return errors.New("stream already closed")
	}

	// Check if the same track has announced already
	if old, ok := sas.announced[suffix]; ok {
		if old.AnnounceStatus == message.ENDED {
			return fmt.Errorf("ended announcement already set: %s", suffix)
		}
	}

	// Check if the same track is to announce
	if old, ok := sas.pending[suffix]; ok {
		if old.AnnounceStatus == message.ENDED {
			// Skip
			slog.Warn("ended announcement already set", "track_suffix", suffix)
			return nil
		}
	}

	am := message.AnnounceMessage{
		AnnounceStatus: message.ENDED,
		TrackSuffix:    suffix,
	}

	sas.pending[suffix] = am

	slog.Debug("set ended announcement", slog.String("track_suffix", suffix))

	return nil
}

func (sas *SendAnnounceStream) SendAnnouncements() error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return fmt.Errorf("stream already closed due to: %w", sas.closeErr)
		}
		return errors.New("stream already closed")
	}

	if len(sas.announced) == 0 {
		return nil
	}

	// Calculate the total length of the ANNOUNCE messages
	var totalLen int
	for _, am := range sas.announced {
		totalLen += am.Len()
	}
	live := message.AnnounceMessage{
		AnnounceStatus: message.LIVE,
	}
	totalLen += live.Len()

	// Create a buffer
	buf := bytes.NewBuffer(make([]byte, 0, totalLen))

	// Encode the ANNOUNCE messages
	for _, am := range sas.announced {
		// Encode the ANNOUNCE message
		_, err := am.Encode(buf)
		if err != nil {
			slog.Error("failed to encode an ANNOUNCE message", "error", err)
			return err
		}
	}
	// Encode the LIVE message
	_, err := live.Encode(buf)
	if err != nil {
		slog.Error("failed to encode a LIVE message", "error", err)
		return err
	}

	// Send the ANNOUNCE messages
	_, err = sas.Stream.Write(buf.Bytes())
	if err != nil {
		slog.Error("failed to send ANNOUNCE messages", "error", err)
		return err
	}

	slog.Debug("sent announcement messages successfully", slog.Any("stream_id", sas.Stream.StreamID()))

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

	sas.closed = true

	slog.Debug("closed a send announce stream gracefully",
		slog.Any("stream_id", sas.Stream.StreamID()),
	)

	return sas.Stream.Close()
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

	slog.Debug("closed a send announce stream with an error",
		slog.Any("stream_id", sas.Stream.StreamID()),
		slog.String("reason", err.Error()),
	)

	return nil
}
