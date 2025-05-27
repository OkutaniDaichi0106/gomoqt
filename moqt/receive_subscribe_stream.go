package moqt

import (
	"context"
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type ReceiveSubscribeStream interface {
	SubscribeID() SubscribeID
	SubuscribeConfig() *SubscribeConfig
	Updated() <-chan struct{}
	Done() <-chan struct{}
}

func newReceiveSubscribeStream(trackCtx *trackContext, stream quic.Stream, config *SubscribeConfig) *receiveSubscribeStream {
	rss := &receiveSubscribeStream{
		config:    config,
		stream:    stream,
		updatedCh: make(chan struct{}, 1),
	}

	go rss.listenUpdates()

	return rss
}

var _ ReceiveSubscribeStream = (*receiveSubscribeStream)(nil)

type receiveSubscribeStream struct {
	trackCtx *trackContext

	stream quic.Stream
	mu     sync.Mutex

	config    *SubscribeConfig
	updatedCh chan struct{}
}

func (rss *receiveSubscribeStream) SubscribeID() SubscribeID {
	return rss.trackCtx.id
}

// func (rss *receiveSubscribeStream) TrackPath() BroadcastPath {
// 	return rss.path
// }

func (rss *receiveSubscribeStream) SubuscribeConfig() *SubscribeConfig {
	return rss.config
}

func (rss *receiveSubscribeStream) Updated() <-chan struct{} {
	return rss.updatedCh
}

func (rss *receiveSubscribeStream) Done() <-chan struct{} {
	return rss.trackCtx.Done()
}

func (rss *receiveSubscribeStream) listenUpdates() {
	var sum message.SubscribeUpdateMessage

	for {
		if rss.trackCtx.Err() != nil {
			if logger := rss.trackCtx.Logger(); logger != nil {
				logger.Error("cancel listening SUBSCRIBE_UPDATE messages",
					"reason", context.Cause(rss.trackCtx),
				)
			}
			return
		}

		_, err := sum.Decode(rss.stream)
		if err != nil {
			if logger := rss.trackCtx.Logger(); logger != nil {
				logger.Error("failed to receive a SUBSCRIBE_UPDATE message",
					"error", err,
				)
			}
			return
		}

		if logger := rss.trackCtx.Logger(); logger != nil {
			logger.Debug("received a SUBSCRIBE_UPDATE message")
		}

		config := &SubscribeConfig{
			TrackPriority:    TrackPriority(sum.TrackPriority),
			MinGroupSequence: GroupSequence(sum.MinGroupSequence),
			MaxGroupSequence: GroupSequence(sum.MaxGroupSequence),
		}

		rss.config = config

		select {
		case rss.updatedCh <- struct{}{}:
		default:
		}

		if logger := rss.trackCtx.Logger(); logger != nil {
			logger.Debug("")
		}
	}
}

func (rss *receiveSubscribeStream) close() error {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if err := rss.closedErr(); err != nil {
		return err
	}

	err := rss.stream.Close()
	if err != nil {
		if logger := rss.trackCtx.Logger(); logger != nil {
			logger.Error("failed to close a receive subscribe stream",
				"error", err,
			)
		}
	}

	if logger := rss.trackCtx.Logger(); logger != nil {
		logger.Debug("closed a receive subscribe stream")
	}

	close(rss.updatedCh)

	return nil
}

func (rss *receiveSubscribeStream) closeWithError(reason error) error {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if err := rss.closedErr(); err != nil {
		return err
	}

	rss.trackCtx.cancel(reason)

	if reason == nil {
		reason = ErrInternalError
	}

	var suberr SubscribeError
	if !errors.As(reason, &suberr) {
		suberr = ErrInternalError
	}

	code := quic.StreamErrorCode(suberr.SubscribeErrorCode())
	rss.stream.CancelRead(code)
	rss.stream.CancelWrite(code)

	if logger := rss.trackCtx.Logger(); logger != nil {
		logger.Debug("closed a receive subscribe stream with an error",
			"reason", reason,
		)
	}

	close(rss.updatedCh)

	return nil
}

func (rss *receiveSubscribeStream) closedErr() error {
	if err := rss.trackCtx.Err(); err != nil {
		if reason := context.Cause(rss.trackCtx); reason != nil {
			return reason
		}
		return nil
	}

	return nil
}
