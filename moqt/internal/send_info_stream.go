package internal

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newSendInfoStream(stream transport.Stream, imr *message.InfoRequestMessage) *SendInfoStream {
	return &SendInfoStream{
		InfoRequestMessage: *imr,
		Stream:             stream,
	}
}

type SendInfoStream struct {
	InfoRequestMessage message.InfoRequestMessage
	Stream             transport.Stream
	mu                 sync.Mutex
}

func (sis *SendInfoStream) SendInfoAndClose(im message.InfoMessage) error {
	sis.mu.Lock()
	defer sis.mu.Unlock()

	_, err := im.Encode(sis.Stream)
	if err != nil {
		slog.Error("failed to send an INFO message",
			"stream_id", sis.Stream.StreamID(),
			"error", err,
		)
		sis.CloseWithError(err)
		return err
	}

	slog.Debug("sent an INFO message",
		slog.Any("stream_id", sis.Stream.StreamID()),
	)

	sis.Close()

	return nil
}

func (sis *SendInfoStream) CloseWithError(err error) error {
	sis.mu.Lock()
	defer sis.mu.Unlock()

	if err == nil {
		err = ErrInternalError
	}

	var inferr InfoError
	if !errors.As(err, &inferr) {
		err = ErrInternalError.WithReason(err.Error())
	}

	code := transport.StreamErrorCode(inferr.InfoErrorCode())

	sis.Stream.CancelRead(code)
	sis.Stream.CancelWrite(code)

	slog.Debug("closed an info stream with an error",
		slog.String("reason", err.Error()),
		slog.Any("stream_id", sis.Stream.StreamID()),
	)

	return nil
}

func (sis *SendInfoStream) Close() error {
	sis.mu.Lock()
	defer sis.mu.Unlock()

	err := sis.Stream.Close()
	if err != nil {
		slog.Error("failed to close an info stream",
			"stream_id", sis.Stream.StreamID(),
			"error", err,
		)
	}

	slog.Debug("closed an info stream",
		slog.Any("stream_id", sis.Stream.StreamID()),
	)

	return nil
}
