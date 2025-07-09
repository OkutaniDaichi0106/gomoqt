package moqt

import (
	"context"
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSendSubscribeStream(sessCtx context.Context, id SubscribeID, stream quic.Stream, config *SubscribeConfig) *sendSubscribeStream {
	ctx, cancel := context.WithCancelCause(sessCtx)
	substr := &sendSubscribeStream{
		ctx:    ctx,
		cancel: cancel,
		id:     id,
		// publishAbortedCh: make(chan *SubscribeError),
		config: config,
		stream: stream,
	}

	// go substr.listenAborted()

	return substr
}

var _ SubscribeController = (*sendSubscribeStream)(nil)

type sendSubscribeStream struct {
	ctx    context.Context
	cancel context.CancelCauseFunc

	id SubscribeID

	config *SubscribeConfig

	stream quic.Stream
	mu     sync.Mutex

	closeErr error
	closed   bool
}

func (sss *sendSubscribeStream) SubscribeID() SubscribeID {
	return sss.id
}

func (sss *sendSubscribeStream) SubscribeConfig() *SubscribeConfig {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	return sss.config
}

func (sss *sendSubscribeStream) UpdateSubscribe(newConfig *SubscribeConfig) error {
	if newConfig == nil {
		return errors.New("new subscribe config cannot be nil")
	}

	sss.mu.Lock()
	defer sss.mu.Unlock()

	if sss.closed {
		if sss.closeErr != nil {
			return sss.closeErr
		}
		return errors.New("stream already closed")
	}

	old := sss.config

	if newConfig.MaxGroupSequence != 0 {
		if newConfig.MinGroupSequence > newConfig.MaxGroupSequence {
			return ErrInvalidRange
		}
	}

	if old.MinGroupSequence != 0 {
		if newConfig.MinGroupSequence == 0 {
			return ErrInvalidRange
		}
		if old.MinGroupSequence > newConfig.MinGroupSequence {
			return ErrInvalidRange
		}
	}
	if old.MaxGroupSequence != 0 {
		if newConfig.MaxGroupSequence == 0 {
			return ErrInvalidRange
		}
		if old.MaxGroupSequence < newConfig.MaxGroupSequence {
			return ErrInvalidRange
		}
	}

	// Send the message first before updating config
	sum := message.SubscribeUpdateMessage{
		TrackPriority:    message.TrackPriority(newConfig.TrackPriority),
		MinGroupSequence: message.GroupSequence(newConfig.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(newConfig.MaxGroupSequence),
	}
	err := sum.Encode(sss.stream)
	if err != nil {
		// Set writeErr and unwritable before closing channel
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			sss.closeErr = &SubscribeError{StreamError: strErr}
		} else {
			strErrCode := quic.StreamErrorCode(InternalSubscribeErrorCode)
			sss.stream.CancelWrite(strErrCode)
			sss.closeErr = &SubscribeError{StreamError: &quic.StreamError{
				StreamID:  sss.stream.StreamID(),
				ErrorCode: strErrCode,
			}}
		}
		sss.closed = true

		sss.cancel(sss.closeErr)

		return sss.closeErr
	}

	// Only update config after successful message sending
	sss.config = newConfig

	return nil
}

func (sss *sendSubscribeStream) Close() error {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	if sss.closed {
		return sss.closeErr
	}

	err := sss.stream.Close()
	if err != nil {
		return err
	}

	sss.closed = true

	sss.cancel(nil)

	return nil
}

func (sss *sendSubscribeStream) CloseWithError(code SubscribeErrorCode) error {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	if sss.closed {
		return sss.closeErr
	}

	strErrCode := quic.StreamErrorCode(code)
	sss.stream.CancelWrite(strErrCode)

	sss.closed = true
	err := &SubscribeError{
		StreamError: &quic.StreamError{
			StreamID:  sss.stream.StreamID(),
			ErrorCode: strErrCode,
		},
	}
	sss.closeErr = err

	sss.stream.CancelRead(strErrCode)

	sss.cancel(err)

	return nil
}
