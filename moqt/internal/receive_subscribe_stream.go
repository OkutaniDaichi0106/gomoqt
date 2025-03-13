package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newReceiveSubscribeStream(sm *message.SubscribeMessage, stream transport.Stream) *ReceiveSubscribeStream {
	rss := &ReceiveSubscribeStream{
		SubscribeID:      (*sm).SubscribeID,
		TrackPath:        (*sm).TrackPath,
		TrackPriority:    (*sm).TrackPriority,
		GroupOrder:       (*sm).GroupOrder,
		MinGroupSequence: (*sm).MinGroupSequence,
		MaxGroupSequence: (*sm).MaxGroupSequence,
		Stream:           stream,
	}

	return rss
}

type ReceiveSubscribeStream struct {
	SubscribeID      message.SubscribeID
	TrackPath        string
	TrackPriority    message.TrackPriority
	GroupOrder       message.GroupOrder
	MinGroupSequence message.GroupSequence
	MaxGroupSequence message.GroupSequence
	Stream           transport.Stream
	mu               sync.Mutex

	closed   bool
	closeErr error
}

// TODO: Implement this method
// func (rss *ReceiveSubscribeStream) SendSubscribeGap(sgm message.SubscribeGapMessage) error {
// 	slog.Debug("sending a data gap", slog.Any("gap", sgm))

// 	rss.mu.Lock()
// 	defer rss.mu.Unlock()

// 	_, err := sgm.Encode(rss.Stream)
// 	if err != nil {
// 		slog.Error("failed to write a subscribe gap message", "error", err,
// 		return err
// 	}

// 	slog.Debug("sent a data gap", slog.Any("gap", sgm))

// 	return nil
// }

func (rss *ReceiveSubscribeStream) ReceiveSubscribeUpdate(sum *message.SubscribeUpdateMessage) error {
	_, err := sum.Decode(rss.Stream)
	if err != nil {
		slog.Error("failed to receive a SUBSCRIBE_UPDATE message",
			"stream_id", rss.Stream.StreamID(),
			"subscribe_id", rss.SubscribeID,
			"error", err,
		)
		return err
	}

	slog.Debug("received a SUBSCRIBE_UPDATE message",
		"stream_id", rss.Stream.StreamID(),
		"subscribe_id", rss.SubscribeID,
	)

	return nil
}

func (rss *ReceiveSubscribeStream) CloseWithError(err error) error {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if rss.closed {
		if rss.closeErr != nil {
			return fmt.Errorf("stream has already closed due to: %w", rss.closeErr)
		}
		return errors.New("stream has already closed")
	}

	rss.closed = true

	if err == nil {
		err = ErrInternalError
	}

	var suberr SubscribeError
	if !errors.As(err, &suberr) {
		suberr = ErrInternalError
	}

	code := transport.StreamErrorCode(suberr.SubscribeErrorCode())
	rss.Stream.CancelRead(code)
	rss.Stream.CancelWrite(code)

	slog.Debug("closed a receive subscribe stream with an error",
		"reason", err,
		"stream_id", rss.Stream.StreamID(),
	)

	return nil
}

func (rss *ReceiveSubscribeStream) Close() error {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if rss.closed {
		if rss.closeErr != nil {
			return fmt.Errorf("stream has already closed due to: %w", rss.closeErr)
		}
		return errors.New("stream has already closed")
	}

	rss.closed = true

	err := rss.Stream.Close()
	if err != nil {
		slog.Error("failed to close a receive subscribe stream",
			"stream_id", rss.Stream.StreamID(),
			"error", err,
		)
	}

	slog.Debug("closed a receive subscribe stream",
		"stream_id", rss.Stream.StreamID(),
	)

	return nil
}
