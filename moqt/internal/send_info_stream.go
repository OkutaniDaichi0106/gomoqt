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

func (sis *SendInfoStream) SendInfoAndClose(i message.InfoMessage) error {
	sis.mu.Lock()
	defer sis.mu.Unlock()

	im := message.InfoMessage{
		TrackPriority:       message.TrackPriority(i.TrackPriority),
		LatestGroupSequence: message.GroupSequence(i.LatestGroupSequence),
		GroupOrder:          message.GroupOrder(i.GroupOrder),
	}

	_, err := im.Encode(sis.Stream)
	if err != nil {
		slog.Error("failed to send an INFO message", slog.String("error", err.Error()))
		return err
	}

	slog.Info("sended an info")

	sis.Close()

	return nil
}

func (sis *SendInfoStream) CloseWithError(err error) error {
	sis.mu.Lock()
	defer sis.mu.Unlock()

	if err == nil {
		return sis.Close()
	}

	sis.mu.Lock()
	defer sis.mu.Unlock()

	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		inferr, ok := err.(InfoError)
		if ok {
			code = transport.StreamErrorCode(inferr.InfoErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	sis.Stream.CancelRead(code)
	sis.Stream.CancelWrite(code)

	slog.Info("rejected an info request")

	return nil
}

func (sis *SendInfoStream) Close() error {
	sis.mu.Lock()
	defer sis.mu.Unlock()

	return sis.Stream.Close()
}
