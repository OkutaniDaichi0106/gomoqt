package moqt

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSendSubscribeStream(id SubscribeID, path TrackPath, config *SubscribeConfig, stream quic.Stream) *sendSubscribeStream {
	return &sendSubscribeStream{
		id:     id,
		path:   path,
		config: config,
		stream: stream,
	}
}

type sendSubscribeStream struct {
	id     SubscribeID
	path   TrackPath
	config *SubscribeConfig

	stream quic.Stream
	mu     sync.Mutex
}

func (sss *sendSubscribeStream) SubscribeID() SubscribeID {
	return sss.id
}

func (sss *sendSubscribeStream) TrackPath() TrackPath {
	return sss.path
}

func (sss *sendSubscribeStream) SubuscribeConfig() *SubscribeConfig {
	return sss.config
}

func (sss *sendSubscribeStream) SendSubscribeUpdate(sum message.SubscribeUpdateMessage) error {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	if sum.MinGroupSequence > sum.MaxGroupSequence {
		return ErrInvalidRange
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

	return nil
}

// func (ss *SendSubscribeStream) ReceiveSubscribeGap() (message.SubscribeGapMessage, error) {
// 	slog.Debug("receiving a data gap")

// 	var gap message.SubscribeGapMessage
// 	_, err := gap.Decode(ss.Stream)
// 	if err != nil {
// 		slog.Error("failed to read a subscribe gap message", "error", err,
// 		return message.SubscribeGapMessage{}, err
// 	}

// 	slog.Debug("received a data gap", slog.Any("gap", gap))

// 	return gap, nil
// }

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
