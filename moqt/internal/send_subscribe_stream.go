package internal

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
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

func (sss *SendSubscribeStream) SendSubscribeUpdate(sum message.SubscribeUpdateMessage) error {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	if sum.MinGroupSequence > sum.MaxGroupSequence {
		return ErrInvalidRange
	}

	_, err := sum.Encode(sss.Stream)
	if err != nil {
		slog.Error("failed to encode a SUBSCRIBE_UPDATE message",
			"error", err,
			"stream_id", sss.Stream.StreamID(),
		)
		return err
	}

	slog.Debug("sent a subscribe update message",
		"stream_id", sss.Stream.StreamID(),
	)

	return nil
}

// func (ss *SendSubscribeStream) ReceiveSubscribeGap() (message.SubscribeGapMessage, error) {
// 	slog.Debug("receiving a data gap")

// 	var gap message.SubscribeGapMessage
// 	_, err := gap.Decode(ss.Stream)
// 	if err != nil {
// 		slog.Error("failed to read a subscribe gap message", "error", err,
// 		return message.SubscribeGapMessage{}, err
// 	}

// 	slog.Debug("received a data gap", slog.Any("gap", gap))

// 	return gap, nil
// }

func (ss *SendSubscribeStream) Close() error {
	err := ss.Stream.Close()
	if err != nil {
		slog.Debug("failed to close a subscrbe send stream",
			"stream_id", ss.Stream.StreamID(),
			"error", err,
		)
		return err
	}

	slog.Debug("closed a subscribe send stream",
		"stream_id", ss.Stream.StreamID(),
	)

	return nil
}

func (sss *SendSubscribeStream) CloseWithError(err error) error {
	if err == nil {
		err = ErrInternalError
	}

	var suberr SubscribeError
	if !errors.As(err, &suberr) {
		suberr = ErrInternalError
	}

	code := transport.StreamErrorCode(suberr.SubscribeErrorCode())

	sss.Stream.CancelRead(code)
	sss.Stream.CancelWrite(code)

	slog.Debug("closed a subscribe receive stream with an error",
		"stream_id", sss.Stream.StreamID(),
		"reason", err.Error(),
	)

	return nil
}
