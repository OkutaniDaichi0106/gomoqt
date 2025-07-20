package moqt

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSendSubscribeStream(id SubscribeID, stream quic.Stream, config *TrackConfig) *sendSubscribeStream {
	ctx, cancel := context.WithCancelCause(stream.Context())
	substr := &sendSubscribeStream{
		ctx:    ctx,
		cancel: cancel,
		id:     id,
		config: config,
		stream: stream,
	}

	return substr
}

type sendSubscribeStream struct {
	ctx    context.Context
	cancel context.CancelCauseFunc

	id SubscribeID

	config *TrackConfig

	stream quic.Stream
	mu     sync.Mutex
}

func (sss *sendSubscribeStream) SubscribeID() SubscribeID {
	return sss.id
}

func (sss *sendSubscribeStream) TrackConfig() *TrackConfig {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	return sss.config
}

func (sss *sendSubscribeStream) UpdateSubscribe(newConfig *TrackConfig) error {
	if newConfig == nil {
		return errors.New("new subscribe config cannot be nil")
	}

	sss.mu.Lock()
	defer sss.mu.Unlock()

	if err := sss.ctx.Err(); err != nil {
		reason := context.Cause(sss.ctx)
		if reason != nil {
			err = reason
		}
		return fmt.Errorf("subscribe stream is closed: %w", err)
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
			err = &SubscribeError{StreamError: strErr}
		} else {
			strErrCode := quic.StreamErrorCode(InternalSubscribeErrorCode)
			sss.stream.CancelWrite(strErrCode)
			err = &SubscribeError{StreamError: &quic.StreamError{
				StreamID:  sss.stream.StreamID(),
				ErrorCode: strErrCode,
			}}
		}

		sss.cancel(err)

		return fmt.Errorf("failed to send subscribe update message: %w", err)
	}

	// Only update config after successful message sending
	sss.config = newConfig

	return nil
}

func (sss *sendSubscribeStream) close() error {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	if err := sss.ctx.Err(); err != nil {
		return nil
	}

	err := sss.stream.Close()
	if err != nil {
		return err
	}

	sss.cancel(nil)

	return nil
}

func (sss *sendSubscribeStream) closeWithError(code SubscribeErrorCode) error {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	if err := sss.ctx.Err(); err != nil {
		return nil
	}

	strErrCode := quic.StreamErrorCode(code)
	sss.stream.CancelWrite(strErrCode)

	err := &SubscribeError{
		StreamError: &quic.StreamError{
			StreamID:  sss.stream.StreamID(),
			ErrorCode: strErrCode,
		},
	}

	sss.stream.CancelRead(strErrCode)

	sss.cancel(err)

	return nil
}
