package moqt

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSendInfoStream(stream quic.Stream, path TrackPath) *sendInfoStream {
	return &sendInfoStream{
		path:   path,
		stream: stream,
	}
}

type sendInfoStream struct {
	path   TrackPath
	stream quic.Stream
	mu     sync.Mutex
}

func (sis *sendInfoStream) SendInfoAndClose(im message.InfoMessage) error {
	sis.mu.Lock()
	defer sis.mu.Unlock()

	_, err := im.Encode(sis.stream)
	if err != nil {
		slog.Error("failed to send an INFO message",
			"stream_id", sis.stream.StreamID(),
			"error", err,
		)
		sis.CloseWithError(err)
		return err
	}

	slog.Debug("sent an INFO message",
		slog.Any("stream_id", sis.stream.StreamID()),
	)

	sis.Close()

	return nil
}

func (sis *sendInfoStream) CloseWithError(err error) error {
	sis.mu.Lock()
	defer sis.mu.Unlock()

	if err == nil {
		err = ErrInternalError
	}

	var inferr InfoError
	if !errors.As(err, &inferr) {
		err = ErrInternalError.WithReason(err.Error())
	}

	code := quic.StreamErrorCode(inferr.InfoErrorCode())

	sis.stream.CancelRead(code)
	sis.stream.CancelWrite(code)

	slog.Debug("closed an info stream with an error",
		slog.String("reason", err.Error()),
		slog.Any("stream_id", sis.stream.StreamID()),
	)

	return nil
}

func (sis *sendInfoStream) Close() error {
	sis.mu.Lock()
	defer sis.mu.Unlock()

	err := sis.stream.Close()
	if err != nil {
		slog.Error("failed to close an info stream",
			"stream_id", sis.stream.StreamID(),
			"error", err,
		)
	}

	slog.Debug("closed an info stream",
		slog.Any("stream_id", sis.stream.StreamID()),
	)

	return nil
}
