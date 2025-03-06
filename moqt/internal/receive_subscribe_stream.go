package internal

import (
	"errors"
	"fmt"
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
// 		slog.Error("failed to write a subscribe gap message", slog.String("error", err.Error()))
// 		return err
// 	}

// 	slog.Debug("sent a data gap", slog.Any("gap", sgm))

// 	return nil
// }

func (rss *ReceiveSubscribeStream) ReceiveSubscribeUpdate(sum *message.SubscribeUpdateMessage) error {
	_, err := sum.Decode(rss.Stream)
	return err
}

func (srs *ReceiveSubscribeStream) CloseWithError(err error) error {
	srs.mu.Lock()
	defer srs.mu.Unlock()

	if err == nil {
		err = ErrInternalError
	}

	var suberr SubscribeError
	if !errors.As(err, &suberr) {
		suberr = ErrInternalError
	}

	code := transport.StreamErrorCode(suberr.SubscribeErrorCode())
	srs.Stream.CancelRead(code)
	srs.Stream.CancelWrite(code)

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
	return rss.Stream.Close()
}
