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
		id:        id,
		config:    config,
		stream:    stream,
		updatedCh: make(chan struct{}, 1),
	}

	go rss.listenUpdates()

	return rss
}

var _ ReceiveSubscribeStream = (*receiveSubscribeStream)(nil)

type receiveSubscribeStream struct {
	id SubscribeID

	stream quic.Stream

	configMu   sync.Mutex
	config     *SubscribeConfig
	updatedCh  chan struct{}
	listenOnce sync.Once

	subscribeCanceledCh chan *SubscribeError

	unwritable bool
	writeErr   error

	unreadable bool
	readErr    error

	closed  bool       // Track if the channel is closed
	closeMu sync.Mutex // Protect against concurrent close operations
}

func (rss *receiveSubscribeStream) SubscribeID() SubscribeID {
	return rss.id
}

func (rss *receiveSubscribeStream) SubscribeConfig() (*SubscribeConfig, error) {
	rss.configMu.Lock()
	defer rss.configMu.Unlock()

	if rss.unreadable {
		if rss.readErr != nil {
			return nil, rss.readErr
		}
		return nil, io.EOF
	}

	return rss.config, nil
}

func (rss *receiveSubscribeStream) Updated() <-chan struct{} {
	return rss.updatedCh
}

func (rss *receiveSubscribeStream) listenUpdates() {
	rss.listenOnce.Do(func() {
		var sum message.SubscribeUpdateMessage
		var err error

		defer func() {
			rss.closeMu.Lock()
			if !rss.closed {
				close(rss.updatedCh)
				rss.closed = true
			}
			rss.closeMu.Unlock()
		}()

		for {
			if rss.unreadable {
				return
			}

			_, err = sum.Decode(rss.stream)
			if err != nil {
				// Check for EOF
				if errors.Is(err, io.EOF) {
					rss.unreadable = true

					return
				}

				// Check for stream error
				var strErr *quic.StreamError
				if errors.As(err, &strErr) {
					rss.unreadable = true
					rss.readErr = &SubscribeError{
						StreamError: strErr,
					}

				} else {
					rss.closeWithError(InternalSubscribeErrorCode)
				}

				select {
				case rss.subscribeCanceledCh <- &SubscribeError{StreamError: strErr}:
				default:
				}

				return
			}

			rss.configMu.Lock()
			rss.config = &SubscribeConfig{
				TrackPriority:    TrackPriority(sum.TrackPriority),
				MinGroupSequence: GroupSequence(sum.MinGroupSequence),
				MaxGroupSequence: GroupSequence(sum.MaxGroupSequence),
			}
			rss.configMu.Unlock()

			select {
			case rss.updatedCh <- struct{}{}:
			default:
			}
		}
	})
}

func (rss *receiveSubscribeStream) done() (error, bool) {
	return rss.writeErr, rss.unreadable && rss.unwritable
}

func (rss *receiveSubscribeStream) cancelWrite(code SubscribeErrorCode) {
	// Close send side of the stream
	if !rss.unwritable {
		strErrCode := quic.StreamErrorCode(code)
		rss.stream.CancelWrite(strErrCode)

		rss.unwritable = true
		rss.writeErr = &SubscribeError{
			StreamError: &quic.StreamError{
				StreamID:  rss.stream.StreamID(),
				ErrorCode: strErrCode,
			},
		}
	}
}

func (rss *receiveSubscribeStream) cancelRead(code SubscribeErrorCode) {
	if !rss.unreadable {
		strErrCode := quic.StreamErrorCode(code)
		rss.stream.CancelRead(strErrCode)

		rss.unreadable = true
		rss.readErr = &SubscribeError{
			StreamError: &quic.StreamError{
				StreamID:  rss.stream.StreamID(),
				ErrorCode: strErrCode,
			},
		}
	}
}

func (rss *receiveSubscribeStream) close() error {
	err, ok := rss.done()
	if ok {
		return err
	}

	// Close send side of the stream
	if !rss.unwritable {
		err = rss.stream.Close()
		if err != nil {
			return err
		}

		rss.unwritable = true
	}

	rss.cancelRead(InternalSubscribeErrorCode)

	rss.closeMu.Lock()
	if !rss.closed {
		close(rss.updatedCh)
		rss.closed = true
	}
	rss.closeMu.Unlock()

	return nil
}

func (rss *receiveSubscribeStream) closeWithError(code SubscribeErrorCode) error {
	err, ok := rss.done()
	if ok {
		return err
	}

	rss.cancelWrite(code)

	rss.cancelRead(code)

	rss.closeMu.Lock()
	if !rss.closed {
		close(rss.updatedCh)
		rss.closed = true
	}
	rss.closeMu.Unlock()

	return nil
}
