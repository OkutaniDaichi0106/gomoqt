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

	configMu   sync.Mutex
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
		rss.configMu.Lock()
		defer rss.configMu.Unlock()
		if rss.ctx.Err() != nil {
			err = context.Cause(rss.ctx)
			return
		}
		sum := message.SubscribeOkMessage{
			GroupOrder: message.GroupOrder(info.GroupOrder),
		}
		err = sum.Encode(rss.stream)
		if err != nil {
			rss.closeWithError(InternalSubscribeErrorCode)
			return
		}
	})

	return err
}

func (rss *receiveSubscribeStream) TrackConfig() *TrackConfig {
	rss.configMu.Lock()
	defer rss.configMu.Unlock()

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
			rss.configMu.Lock()
			if rss.ctx.Err() != nil {
				rss.configMu.Unlock()
				break
			}
			rss.configMu.Unlock()

			err = sum.Decode(rss.stream)
			if err != nil {
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
				break
			}

			rss.configMu.Lock()
			rss.config = &TrackConfig{
				TrackPriority:    sum.TrackPriority,
				MinGroupSequence: sum.MinGroupSequence,
				MaxGroupSequence: sum.MaxGroupSequence,
			}

			select {
			case rss.updatedCh <- struct{}{}:
			default:
			}
			rss.configMu.Unlock()
		}

		// // Cleanup after loop ends
		// rss.configMu.Lock()
		// // Always close the channel if it hasn't been closed yet
		// select {
		// case <-rss.updatedCh:
		// 	// Channel is already closed
		// default:
		// 	close(rss.updatedCh)
		// }

		// rss.configMu.Unlock()
	})
}

func (rss *receiveSubscribeStream) close() error {
	rss.configMu.Lock()
	defer rss.configMu.Unlock()

	if rss.ctx.Err() != nil {
		return context.Cause(rss.ctx)
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
	rss.configMu.Lock()
	defer rss.configMu.Unlock()

	if rss.ctx.Err() != nil {
		return context.Cause(rss.ctx)
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
