package internal

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newReceiveSubscribeStream(sm *message.SubscribeMessage, stream transport.Stream) *ReceiveSubscribeStream {
	rss := &ReceiveSubscribeStream{
		SubscribeMessage: *sm,
		Stream:           stream,
	}

	go rss.listenSubscribeUpdates()

	return rss
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
	var code SubscribeErrorCode

	var suberr SubscribeError
	if errors.As(err, &suberr) {
		code = suberr.SubscribeErrorCode()
	} else {
		code = ErrInternalError.SubscribeErrorCode()
	}

	srs.Stream.CancelRead(transport.StreamErrorCode(code))
	srs.Stream.CancelWrite(transport.StreamErrorCode(code))

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

func (rss *ReceiveSubscribeStream) listenSubscribeUpdates() {
	var sum message.SubscribeUpdateMessage
	for {
		_, err := sum.Decode(rss.Stream)
		if err != nil {
			slog.Error("failed to read a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
			rss.CloseWithError(err)
			break
		}

		rss.mu.Lock()
		updateSubscription(&rss.SubscribeMessage, &sum)
		rss.mu.Unlock()
	}
}
