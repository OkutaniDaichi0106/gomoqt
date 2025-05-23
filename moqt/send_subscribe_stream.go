package moqt

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type sentSubscription struct {
	groupQueue      *incomingGroupStreamQueue
	subscribeStream *sendSubscribeStream
}

func newSendSubscribeStream(id SubscribeID, path BroadcastPath, config *SubscribeConfig, stream quic.Stream) *sendSubscribeStream {
	return &sendSubscribeStream{
		id:     id,
		path:   path,
		config: config,
		stream: stream,
	}
}

var _ SentSubscription = (*sendSubscribeStream)(nil)

type sendSubscribeStream struct {
	id     SubscribeID
	path   BroadcastPath
	config *SubscribeConfig

	stream quic.Stream
	mu     sync.Mutex
}

func (sss *sendSubscribeStream) SubscribeID() SubscribeID {
	return sss.id
}

func (sss *sendSubscribeStream) TrackPath() BroadcastPath {
	return sss.path
}

func (sss *sendSubscribeStream) SubuscribeConfig() *SubscribeConfig {
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
		GroupOrder:       message.GroupOrder(new.GroupOrder),
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

func (ss *sendSubscribeStream) Close() error {
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

func (sss *sendSubscribeStream) CloseWithError(err error) error {
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
