package internal

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newReceiveSubscribeStream(sm *message.SubscribeMessage, stream transport.Stream) *ReceiveSubscribeStream {
	return &ReceiveSubscribeStream{
		SubscribeMessage: *sm,
		Stream:           stream,
	}
}

type ReceiveSubscribeStream struct {
	SubscribeMessage message.SubscribeMessage
	Stream           transport.Stream
	mu               sync.Mutex
}

func (rss *ReceiveSubscribeStream) SendSubscribeGap(sgm message.SubscribeGapMessage) error {
	slog.Debug("sending a data gap", slog.Any("gap", sgm))

	rss.mu.Lock()
	defer rss.mu.Unlock()

	_, err := sgm.Encode(rss.Stream)
	if err != nil {
		slog.Error("failed to write a subscribe gap message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("sent a data gap", slog.Any("gap", sgm))

	return nil
}

func (srs *ReceiveSubscribeStream) CloseWithError(err error) error {
	slog.Debug("closing a subscrbe receive stream", slog.Any("subscription", srs.SubscribeMessage))

	if err == nil {
		return srs.Close()
	}

	// TODO:

	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		feterr, ok := err.(FetchError)
		if ok {
			code = transport.StreamErrorCode(feterr.FetchErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	srs.Stream.CancelRead(code)
	srs.Stream.CancelWrite(code)

	slog.Debug("closed a subscrbe receive stream", slog.Any("subscription", srs.SubscribeMessage))

	return nil
}

func (srs *ReceiveSubscribeStream) Close() error {
	slog.Debug("closing a subscrbe receive stream", slog.Any("subscription", srs.SubscribeMessage))

	err := srs.Stream.Close()
	if err != nil {
		slog.Debug("catch an error when closing a Subscribe Stream", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("closed a subscrbe receive stream", slog.Any("subscription", srs.SubscribeMessage))

	return nil
}
