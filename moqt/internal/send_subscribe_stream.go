package internal

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newSendSubscribeStream(sm *message.SubscribeMessage, stream transport.Stream) *SendSubscribeStream {
	return &SendSubscribeStream{
		SubscribeMessage: *sm,
		Stream:           stream,
	}
}

type SendSubscribeStream struct {
	SubscribeMessage message.SubscribeMessage
	Stream           transport.Stream
	mu               sync.Mutex
}

func (sss *SendSubscribeStream) UpdateSubscribe(sum message.SubscribeUpdateMessage) error {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	if sum.MinGroupSequence > sum.MaxGroupSequence {
		return ErrInvalidRange
	}

	_, err := sum.Encode(sss.Stream)
	if err != nil {
		slog.Error("failed to write a subscribe update message", slog.String("error", err.Error()))
		return err
	}

	// Update the SubscribeMessage
	updateSubscription(&sss.SubscribeMessage, &sum)

	slog.Debug("updated a subscription", slog.Any("subscription", sss.SubscribeMessage))

	return nil
}

func (ss *SendSubscribeStream) ReceiveSubscribeGap() (message.SubscribeGapMessage, error) {
	slog.Debug("receiving a data gap")

	var gap message.SubscribeGapMessage
	_, err := gap.Decode(ss.Stream)
	if err != nil {
		slog.Error("failed to read a subscribe gap message", slog.String("error", err.Error()))
		return message.SubscribeGapMessage{}, err
	}

	slog.Debug("received a data gap", slog.Any("gap", gap))

	return gap, nil
}

func (ss *SendSubscribeStream) Close() error {
	slog.Debug("closing a subscrbe send stream", slog.Any("subscription", ss.SubscribeMessage))

	err := ss.Stream.Close()
	if err != nil {
		slog.Debug("catch an error when closing a Subscribe Stream", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("closed a subscrbe send stream", slog.Any("subscription", ss.SubscribeMessage))

	return nil
}

func (sss *SendSubscribeStream) CloseWithError(err error) error {
	slog.Debug("closing a subscrbe send stream", slog.Any("subscription", sss.SubscribeMessage))

	if err == nil {
		return sss.Close()
	}

	var code protocol.SubscribeErrorCode

	var suberr SubscribeError

	if errors.As(err, &suberr) {
		code = suberr.SubscribeErrorCode()
	} else {
		code = ErrInternalError.SubscribeErrorCode()
	}

	sss.Stream.CancelRead(transport.StreamErrorCode(code))
	sss.Stream.CancelWrite(transport.StreamErrorCode(code))

	slog.Debug("closed a subscrbe receive stream", slog.Any("config", sss.SubscribeMessage))

	return nil
}
