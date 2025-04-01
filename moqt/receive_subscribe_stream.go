package moqt

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newReceiveSubscribeStream(id SubscribeID, path TrackPath, config SubscribeConfig, stream quic.Stream) *receiveSubscribeStream {
	rss := &receiveSubscribeStream{
		id:     id,
		path:   path,
		config: config,
		stream: stream,
	}

	return rss
}

type receiveSubscribeStream struct {
	id     SubscribeID
	path   TrackPath
	config SubscribeConfig
	stream quic.Stream
	mu     sync.Mutex

	closed   bool
	closeErr error
}

// TODO: Implement this method
// func (rss *ReceiveSubscribeStream) SendSubscribeGap(sgm message.SubscribeGapMessage) error {
// 	slog.Debug("sending a data gap", slog.Any("gap", sgm))

// 	rss.mu.Lock()
// 	defer rss.mu.Unlock()

// 	_, err := sgm.Encode(rss.Stream)
// 	if err != nil {
// 		slog.Error("failed to write a subscribe gap message", "error", err,
// 		return err
// 	}

// 	slog.Debug("sent a data gap", slog.Any("gap", sgm))

// 	return nil
// }

func (rss *receiveSubscribeStream) ReceiveSubscribeUpdate(sum *message.SubscribeUpdateMessage) error {
	_, err := sum.Decode(rss.stream)
	if err != nil {
		slog.Error("failed to receive a SUBSCRIBE_UPDATE message",
			"stream_id", rss.stream.StreamID(),
			"subscribe_id", rss.id,
			"error", err,
		)
		return err
	}

	slog.Debug("received a SUBSCRIBE_UPDATE message",
		"stream_id", rss.stream.StreamID(),
		"subscribe_id", rss.id,
	)

	return nil
}

func (rss *receiveSubscribeStream) CloseWithError(err error) error {
	rss.mu.Lock()
	defer rss.mu.Unlock()

	if rss.closed {
		if rss.closeErr != nil {
			return fmt.Errorf("stream has already closed due to: %w", rss.closeErr)
		}
		return errors.New("stream has already closed")
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
			return fmt.Errorf("stream has already closed due to: %w", rss.closeErr)
		}
		return errors.New("stream has already closed")
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
