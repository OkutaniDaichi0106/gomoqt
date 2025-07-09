package moqt

import (
	"context"
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newReceiveSubscribeStream(trackCtx context.Context, id SubscribeID, stream quic.Stream, config *SubscribeConfig) *receiveSubscribeStream {
	ctx, cancel := context.WithCancelCause(trackCtx)
	rss := &receiveSubscribeStream{
		id:        id,
		config:    config,
		stream:    stream,
		updatedCh: make(chan struct{}, 1),
		ctx:       ctx,
		cancel:    cancel,
	}

	go rss.listenUpdates()

	return rss
}

var _ PublishController = (*receiveSubscribeStream)(nil)

type receiveSubscribeStream struct {
	id SubscribeID

	stream quic.Stream

	acceptOnce sync.Once

	mu         sync.Mutex
	config     *SubscribeConfig
	updatedCh  chan struct{}
	listenOnce sync.Once

	ctx    context.Context
	cancel context.CancelCauseFunc
}

func (rss *receiveSubscribeStream) SubscribeID() SubscribeID {
	return rss.id
}

func (rss *receiveSubscribeStream) WriteInfo(info Info) error {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if err := rss.ctx.Err(); err != nil {
		// Return the cause if available, otherwise return the context error
		if cause := context.Cause(rss.ctx); cause != nil {
			return cause
		}
		return err
	}

	rss.accept(info)

	return nil
}

func (rss *receiveSubscribeStream) SubscribeConfig() (*SubscribeConfig, error) {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if err := rss.ctx.Err(); err != nil {
		// Return the cause if available, otherwise return the context error
		if cause := context.Cause(rss.ctx); cause != nil {
			return nil, cause
		}
		return nil, err
	}

	return rss.config, nil
}

func (rss *receiveSubscribeStream) Updated() <-chan struct{} {
	return rss.updatedCh
}

func (rss *receiveSubscribeStream) accept(info Info) {
	rss.acceptOnce.Do(func() {
		sum := message.SubscribeOkMessage{
			GroupOrder: message.GroupOrder(info.GroupOrder),
		}
		err := sum.Encode(rss.stream)
		if err != nil {
			rss.CloseWithError(InternalSubscribeErrorCode)
			return
		}
	})
}

func (rss *receiveSubscribeStream) listenUpdates() {
	rss.listenOnce.Do(func() {
		var sum message.SubscribeUpdateMessage
		var err error

		for {
			rss.mu.Lock()
			if rss.ctx.Err() != nil {
				rss.mu.Unlock()
				break
			}
			rss.mu.Unlock()

			err = sum.Decode(rss.stream)
			if err != nil {
				rss.mu.Lock()
				// Check for stream error
				var strErr *quic.StreamError
				if errors.As(err, &strErr) {
					subErr := &SubscribeError{
						StreamError: strErr,
					}
					rss.cancel(subErr)
				} else {
					rss.cancel(err)
				}
				rss.mu.Unlock()
				break
			}

			rss.mu.Lock()
			rss.config = &SubscribeConfig{
				TrackPriority:    TrackPriority(sum.TrackPriority),
				MinGroupSequence: GroupSequence(sum.MinGroupSequence),
				MaxGroupSequence: GroupSequence(sum.MaxGroupSequence),
			}
			rss.mu.Unlock()

			select {
			case rss.updatedCh <- struct{}{}:
			default:
			}
		}

		// Cleanup after loop ends
		rss.mu.Lock()
		// Always close the channel if it hasn't been closed yet
		select {
		case <-rss.updatedCh:
			// Channel is already closed
		default:
			close(rss.updatedCh)
		}
		rss.mu.Unlock()
	})
}

func (rss *receiveSubscribeStream) Close() error {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if err := rss.ctx.Err(); err != nil {
		// Return the cause if available, otherwise return the context error
		if cause := context.Cause(rss.ctx); cause != nil {
			return cause
		}
		return err
	}

	err := rss.stream.Close()
	rss.cancel(err)

	// TODO: Should we cancel the receive stream here?

	close(rss.updatedCh)

	return err
}

func (rss *receiveSubscribeStream) CloseWithError(code SubscribeErrorCode) error {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if err := rss.ctx.Err(); err != nil {
		// Return the cause if available, otherwise return the context error
		if cause := context.Cause(rss.ctx); cause != nil {
			return cause
		}
		return err
	}

	strErrCode := quic.StreamErrorCode(code)
	rss.stream.CancelWrite(strErrCode)
	rss.stream.CancelRead(strErrCode)

	// Set the close error
	subErr := &SubscribeError{
		StreamError: &quic.StreamError{
			StreamID:  rss.stream.StreamID(),
			ErrorCode: strErrCode,
		},
	}

	rss.cancel(subErr)

	close(rss.updatedCh)

	return nil
}

// func (rss *receiveSubscribeStream) isClosed() (error, bool) {
// 	rss.mu.Lock()
// 	defer rss.mu.Unlock()

// 	return rss.closeErr, rss.closed
// }
