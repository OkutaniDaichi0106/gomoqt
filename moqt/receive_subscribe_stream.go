package moqt

import (
	"context"
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newReceiveSubscribeStream(id SubscribeID, stream quic.Stream, config *TrackConfig) *receiveSubscribeStream {
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() {
		streamCtx := stream.Context()
		<-streamCtx.Done()
		reason := context.Cause(streamCtx)
		var (
			strErr *quic.StreamError
			appErr *quic.ApplicationError
		)
		if errors.As(reason, &strErr) {
			reason = &SubscribeError{
				StreamError: strErr,
			}
		} else if errors.As(reason, &appErr) {
			reason = &SessionError{
				ApplicationError: appErr,
			}
		}
		cancel(reason)
	}()

	rss := &receiveSubscribeStream{
		subscribeID: id,
		config:      config,
		stream:      stream,
		updatedCh:   make(chan struct{}, 1),
		ctx:         ctx,
		cancel:      cancel,
	}

	go rss.listenUpdates()

	return rss
}

type receiveSubscribeStream struct {
	subscribeID SubscribeID

	stream quic.Stream

	acceptOnce sync.Once

	mu         sync.Mutex
	config     *TrackConfig
	updatedCh  chan struct{}
	listenOnce sync.Once

	ctx    context.Context
	cancel context.CancelCauseFunc
}

func (rss *receiveSubscribeStream) SubscribeID() SubscribeID {
	return rss.subscribeID
}

func (rss *receiveSubscribeStream) writeInfo(info Info) error {
	var err error
	rss.acceptOnce.Do(func() {
		rss.mu.Lock()
		defer rss.mu.Unlock()
		if err = rss.ctx.Err(); err != nil {
			// Return the cause if available, otherwise return the context error
			if cause := context.Cause(rss.ctx); cause != nil {
				err = cause
				return
			}
		}
		sum := message.SubscribeOkMessage{
			GroupOrder: message.GroupOrder(info.GroupOrder),
		}
		err := sum.Encode(rss.stream)
		if err != nil {
			rss.closeWithError(InternalSubscribeErrorCode)
			return
		}
	})

	return err
}

func (rss *receiveSubscribeStream) TrackConfig() *TrackConfig {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	return rss.config
}

func (rss *receiveSubscribeStream) Updated() <-chan struct{} {
	return rss.updatedCh
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
			rss.config = &TrackConfig{
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

func (rss *receiveSubscribeStream) close() error {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if err := rss.ctx.Err(); err != nil {
		// Return the cause if available, otherwise return the context error
		if cause := context.Cause(rss.ctx); cause != nil {
			return cause
		}
		return err
	}

	// Close the write-side stream
	err := rss.stream.Close()
	// Cancel the read-side stream
	rss.stream.CancelRead(quic.StreamErrorCode(PublishAbortedErrorCode))

	rss.cancel(nil)

	close(rss.updatedCh)

	return err
}

func (rss *receiveSubscribeStream) closeWithError(code SubscribeErrorCode) error {
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
	// Cancel the write-side stream
	rss.stream.CancelWrite(strErrCode)
	// Cancel the read-side stream
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
