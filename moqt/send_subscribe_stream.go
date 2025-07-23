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
	substr := &sendSubscribeStream{
		id:     id,
		config: config,
		stream: stream,
	}

	return substr
}

type sendSubscribeStream struct {
	// closed bool

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

	if sss.stream.Context().Err() != nil {
		return context.Cause(sss.stream.Context())
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

		return fmt.Errorf("failed to send subscribe update message: %w", err)
	}

	// Only update config after successful message sending
	sss.config = newConfig

	return nil
}

func (sss *sendSubscribeStream) Context() context.Context {
	return sss.stream.Context()
}

func (sss *sendSubscribeStream) close() error {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	// Close the write side of the stream
	err := sss.stream.Close()
	// Cancel the read side of the stream
	strErrCode := quic.StreamErrorCode(SubscribeCanceledErrorCode)
	sss.stream.CancelRead(strErrCode)

	return err
}

func (sss *sendSubscribeStream) closeWithError(code SubscribeErrorCode) error {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	strErrCode := quic.StreamErrorCode(code)
	// Cancel the write side of the stream
	sss.stream.CancelWrite(strErrCode)
	// Cancel the read side of the stream
	sss.stream.CancelRead(strErrCode)

	return nil
}
