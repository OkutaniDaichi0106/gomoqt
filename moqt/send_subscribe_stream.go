package moqt

import (
	"context"
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

func newSendSubscribeStream(id SubscribeID, stream quic.Stream, initConfig *TrackConfig) *sendSubscribeStream {
	substr := &sendSubscribeStream{
		ctx:    context.WithValue(stream.Context(), &biStreamTypeCtxKey, message.StreamTypeSubscribe),
		id:     id,
		config: initConfig,
		stream: stream,
	}

	return substr
}

type sendSubscribeStream struct {
	ctx context.Context

	config *TrackConfig

	stream quic.Stream

	mu sync.Mutex

	id SubscribeID
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
		return errors.New("new track config cannot be nil")
	}

	sss.mu.Lock()
	defer sss.mu.Unlock()

	if sss.ctx.Err() != nil {
		return Cause(sss.ctx)
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
		// Close the stream with error on write failure
		sss.mu.Unlock() // Unlock before calling closeWithError to avoid deadlock
		sss.closeWithError(InternalSubscribeErrorCode)
		sss.mu.Lock() // Re-lock for defer
		return err
	}

	// Only update config after successful message sending
	sss.config = newConfig

	return nil
}

func (sss *sendSubscribeStream) Context() context.Context {
	return sss.ctx
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
