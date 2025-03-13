package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newReceiveAnnounceStream(apm *message.AnnouncePleaseMessage, stream transport.Stream) *ReceiveAnnounceStream {
	ras := &ReceiveAnnounceStream{
		AnnouncePleaseMessage: *apm,
		stream:                stream,
	}

	return ras
}

type ReceiveAnnounceStream struct {
	AnnouncePleaseMessage message.AnnouncePleaseMessage
	stream                transport.Stream

	mu       sync.RWMutex
	closed   bool
	closeErr error
}

func (ras *ReceiveAnnounceStream) ReceiveAnnounceMessage(am *message.AnnounceMessage) error {
	_, err := am.Decode(ras.stream)
	if err != nil {
		slog.Error("failed to decode an ANNOUNCE message", "error", err)
		return err
	}

	slog.Debug("received an ANNOUNCE message", slog.Any("announce", am))

	return nil
}

func (ras *ReceiveAnnounceStream) Close() error {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	if ras.closed {
		if ras.closeErr == nil {
			return fmt.Errorf("stream has already closed due to: %v", ras.closeErr)
		}

		return errors.New("stream has already closed")
	}

	ras.closed = true
	err := ras.stream.Close()
	if err != nil {
		return err
	}

	slog.Debug("closed a receive announce stream",
		slog.Any("stream_id", ras.stream.StreamID()),
	)

	return nil
}

func (ras *ReceiveAnnounceStream) CloseWithError(err error) error {
	ras.mu.Lock()
	defer ras.mu.Unlock()

	if ras.closed {
		return ras.closeErr
	}

	if err == nil {
		err = ErrInternalError
	}

	ras.closeErr = err
	ras.closed = true

	var annerr AnnounceError

	if !errors.As(err, &annerr) {
		annerr = ErrInternalError
	}

	code := transport.StreamErrorCode(annerr.AnnounceErrorCode())

	ras.stream.CancelRead(transport.StreamErrorCode(code))
	ras.stream.CancelWrite(transport.StreamErrorCode(code))

	slog.Debug("closed a receive announce stream with an error",
		slog.Any("stream_id", ras.stream.StreamID()),
		slog.String("reason", err.Error()),
	)

	return nil
}
