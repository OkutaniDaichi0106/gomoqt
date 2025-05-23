package moqt

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newReceiveSubscribeStream(id SubscribeID, path BroadcastPath, config *SubscribeConfig, stream quic.Stream) *receiveSubscribeStream {
	rss := &receiveSubscribeStream{
		id:        id,
		path:      path,
		config:    config,
		stream:    stream,
		updatedCh: make(chan struct{}, 1),
	}

	go rss.listenUpdates()

	return rss
}

type receiveSubscribeStream struct {
	id     SubscribeID
	path   BroadcastPath
	config *SubscribeConfig
	stream quic.Stream
	mu     sync.Mutex

	updatedCh chan struct{}

	closed   bool
	closeErr error
}

func (rss *receiveSubscribeStream) NotifyGap(min, max GroupSequence, reason error) error {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	var grperr GroupError
	if !errors.As(reason, grperr) {
		grperr = ErrInternalError
	}

	sgm := message.SubscribeGapMessage{
		StartGroupSequence: message.GroupSequence(min),
		GapCount:           uint64(max - min),
		GroupErrorCode:     grperr.GroupErrorCode(),
	}

	_, err := sgm.Encode(rss.stream)
	if err != nil {
		slog.Error("failed to write a subscribe gap message", "error", err)
		return err
	}

	return nil
}

func (rss *receiveSubscribeStream) SubscribeID() SubscribeID {
	return rss.id
}

func (rss *receiveSubscribeStream) TrackPath() BroadcastPath {
	return rss.path
}

func (rss *receiveSubscribeStream) SubuscribeConfig() *SubscribeConfig {
	return rss.config
}

func (rss *receiveSubscribeStream) Updated() <-chan struct{} {
	return rss.updatedCh
}

func (rss *receiveSubscribeStream) listenUpdates() {
	var sum message.SubscribeUpdateMessage
	for {
		if rss.closed {
			return
		}

		_, err := sum.Decode(rss.stream)
		if err != nil {
			slog.Error("failed to receive a SUBSCRIBE_UPDATE message",
				"stream_id", rss.stream.StreamID(),
				"subscribe_id", rss.id,
				"error", err,
			)
			return
		}

		slog.Debug("received a SUBSCRIBE_UPDATE message",
			"stream_id", rss.stream.StreamID(),
			"subscribe_id", rss.id,
		)

		config := &SubscribeConfig{
			TrackPriority:    TrackPriority(sum.TrackPriority),
			GroupOrder:       GroupOrder(sum.GroupOrder),
			MinGroupSequence: GroupSequence(sum.MinGroupSequence),
			MaxGroupSequence: GroupSequence(sum.MaxGroupSequence),
		}

		rss.config = config

		select {
		case rss.updatedCh <- struct{}{}:
		default:
		}

	}
}

func (rss *receiveSubscribeStream) CloseWithError(err error) error {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if rss.closed {
		if rss.closeErr != nil {
			return rss.closeErr
		}
		return nil
	}

	rss.closed = true

	if err == nil {
		err = ErrInternalError
	}

	var suberr SubscribeError
	if !errors.As(err, &suberr) {
		suberr = ErrInternalError
	}

	code := quic.StreamErrorCode(suberr.SubscribeErrorCode())
	rss.stream.CancelRead(code)
	rss.stream.CancelWrite(code)

	slog.Debug("closed a receive subscribe stream with an error",
		"reason", err,
		"stream_id", rss.stream.StreamID(),
	)

	return nil
}

func (rss *receiveSubscribeStream) Close() error {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if rss.closed {
		if rss.closeErr != nil {
			return rss.closeErr
		}
		return nil
	}

	rss.closed = true

	err := rss.stream.Close()
	if err != nil {
		slog.Error("failed to close a receive subscribe stream",
			"stream_id", rss.stream.StreamID(),
			"error", err,
		)
	}

	slog.Debug("closed a receive subscribe stream",
		"stream_id", rss.stream.StreamID(),
	)

	return nil
}
