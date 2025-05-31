package moqt

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/moqtrace"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type SendSubscribeStream interface {
	SubscribeID() SubscribeID
	SubscribeConfig() *SubscribeConfig
	UpdateSubscribe(*SubscribeConfig) error
}

func newSendSubscribeStream(trackCtx *trackContext, stream quic.Stream, config *SubscribeConfig, streamTracer *moqtrace.StreamTracer) *sendSubscribeStream {
	substr := &sendSubscribeStream{
		trackCtx: trackCtx,
		config:   config,
		stream:   stream,
		tracer:   streamTracer,
	}

	go func() {
		<-trackCtx.Done()
		reason := context.Cause(trackCtx)
		if reason == nil {
			substr.close()
		} else {
			substr.closeWithError(reason)
		}
	}()

	return substr
}

var _ SendSubscribeStream = (*sendSubscribeStream)(nil)

type sendSubscribeStream struct {
	trackCtx *trackContext

	config *SubscribeConfig

	stream quic.Stream
	mu     sync.Mutex

	tracer *moqtrace.StreamTracer
}

func (sss *sendSubscribeStream) SubscribeID() SubscribeID {
	return sss.trackCtx.id
}

func (sss *sendSubscribeStream) SubscribeConfig() *SubscribeConfig {
	return sss.config
}

func (sss *sendSubscribeStream) UpdateSubscribe(new *SubscribeConfig) error {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	old := sss.config

	if new.MaxGroupSequence != 0 {
		if new.MinGroupSequence > new.MaxGroupSequence {
			return ErrInvalidRange
		}
	}

	if old.MinGroupSequence != 0 {
		if new.MinGroupSequence == 0 {
			return ErrInvalidRange
		}
		if old.MinGroupSequence > new.MinGroupSequence {
			return ErrInvalidRange
		}
	}

	if old.MaxGroupSequence != 0 {
		if new.MaxGroupSequence == 0 {
			return ErrInvalidRange
		}
		if old.MaxGroupSequence < new.MaxGroupSequence {
			return ErrInvalidRange
		}
	}

	sum := message.SubscribeUpdateMessage{
		TrackPriority:    message.TrackPriority(new.TrackPriority),
		MinGroupSequence: message.GroupSequence(new.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(new.MaxGroupSequence),
	}

	_, err := sum.Encode(sss.stream)
	if err != nil {
		slog.Error("failed to encode a SUBSCRIBE_UPDATE message",
			"error", err,
			"stream_id", sss.stream.StreamID(),
		)
		return err
	}

	slog.Debug("sent a subscribe update message",
		"stream_id", sss.stream.StreamID(),
	)

	sss.config = new

	return nil
}

func (ss *sendSubscribeStream) close() error {
	err := ss.stream.Close()
	if err != nil {
		slog.Debug("failed to close a subscrbe send stream",
			"stream_id", ss.stream.StreamID(),
			"error", err,
		)
		return err
	}

	slog.Debug("closed a subscribe send stream",
		"stream_id", ss.stream.StreamID(),
	)

	return nil
}

func (sss *sendSubscribeStream) closeWithError(err error) error {
	if err == nil {
		err = ErrInternalError
	}

	var suberr SubscribeError
	if !errors.As(err, &suberr) {
		suberr = ErrInternalError
	}

	code := quic.StreamErrorCode(suberr.SubscribeErrorCode())

	sss.stream.CancelRead(code)
	sss.stream.CancelWrite(code)

	slog.Debug("closed a subscribe receive stream with an error",
		"stream_id", sss.stream.StreamID(),
		"reason", err.Error(),
	)

	return nil
}
