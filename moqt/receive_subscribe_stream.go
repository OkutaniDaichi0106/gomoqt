package moqt

import (
	"errors"
	"io"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type ReceiveSubscribeStream interface {
	SubscribeID() SubscribeID
	SubscribeConfig() (*SubscribeConfig, error)
	Updated() <-chan struct{}
}

func newReceiveSubscribeStream(id SubscribeID, stream quic.Stream, config *SubscribeConfig) *receiveSubscribeStream {
	rss := &receiveSubscribeStream{
		id:                  id,
		config:              config,
		stream:              stream,
		updatedCh:           make(chan struct{}, 1),
		subscribeCanceledCh: make(chan struct{}, 1),
	}

	go rss.listenUpdates()

	return rss
}

var _ ReceiveSubscribeStream = (*receiveSubscribeStream)(nil)

type receiveSubscribeStream struct {
	id SubscribeID

	stream quic.Stream

	acceptOnce sync.Once

	mu         sync.Mutex
	config     *SubscribeConfig
	updatedCh  chan struct{}
	listenOnce sync.Once

	subscribeCanceledCh chan struct{}

	closed   bool // Track if the channel is closed
	closeErr error
}

func (rss *receiveSubscribeStream) SubscribeID() SubscribeID {
	return rss.id
}

func (rss *receiveSubscribeStream) SubscribeConfig() (*SubscribeConfig, error) {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if rss.closed {
		if rss.closeErr != nil {
			return nil, rss.closeErr
		}
		return nil, io.EOF
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
			rss.closeWithError(InternalSubscribeErrorCode)
			return
		}
	})
}

func (rss *receiveSubscribeStream) canceled() <-chan struct{} {
	return rss.subscribeCanceledCh
}

func (rss *receiveSubscribeStream) listenUpdates() {
	rss.listenOnce.Do(func() {
		var sum message.SubscribeUpdateMessage
		var err error

		for {
			rss.mu.Lock()
			if rss.closed {
				rss.mu.Unlock()
				break
			}
			rss.mu.Unlock()

			err = sum.Decode(rss.stream)
			if err != nil {
				rss.mu.Lock()
				if rss.closed {
					rss.mu.Unlock()
					break
				}

				rss.closed = true

				// Check for stream error
				var strErr *quic.StreamError
				if errors.As(err, &strErr) {
					rss.closeErr = &SubscribeError{
						StreamError: strErr,
					}
				} else {
					rss.closeErr = err
				}

				close(rss.subscribeCanceledCh)
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
		if !rss.closed {
			rss.closed = true
		}
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

	if rss.closed {
		return rss.closeErr
	}

	rss.closed = true

	err := rss.stream.Close()

	strErrCode := quic.StreamErrorCode(InternalSubscribeErrorCode)
	rss.stream.CancelRead(strErrCode)

	close(rss.updatedCh)

	return err
}

func (rss *receiveSubscribeStream) closeWithError(code SubscribeErrorCode) error {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if rss.closed {
		return rss.closeErr
	}

	rss.closed = true

	strErrCode := quic.StreamErrorCode(code)
	rss.stream.CancelWrite(strErrCode)
	rss.stream.CancelRead(strErrCode)

	// Set the close error
	rss.closeErr = &SubscribeError{
		StreamError: &quic.StreamError{
			StreamID:  rss.stream.StreamID(),
			ErrorCode: strErrCode,
		},
	}

	close(rss.updatedCh)

	return nil
}

func (rss *receiveSubscribeStream) isClosed() (error, bool) {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	return rss.closeErr, rss.closed
}
