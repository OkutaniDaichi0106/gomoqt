package moqt

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSessionStream(sessCtx *sessionContext, stream quic.Stream) *sessionStream {
	sess := &sessionStream{
		sessCtx: sessCtx,
		stream:  stream,
	}

	// Start listening for updates in a separate goroutine
	go sess.listenUpdates()

	return sess
}

// var _ SessionStream = (*sessionStream)(nil)

type sessionStream struct {
	sessCtx *sessionContext

	stream quic.Stream
	mu     sync.Mutex
}

func (ss *sessionStream) UpdateSession(bitrate uint64) error {
	_, err := message.SessionUpdateMessage{
		Bitrate: bitrate,
	}.Encode(ss.stream)
	if err != nil {
		slog.Error("failed to send a SESSION_UPDATE message", "error", err)
		return err
	}

	slog.Debug("sent a SESSION_UPDATE message")

	return nil
}

func (ss *sessionStream) Close() error {
	if ss.closedErr() != nil {
		return ss.closedErr()
	}

	ss.sessCtx.cancel(nil)

	return ss.stream.Close()
}

func (ss *sessionStream) CloseWithError(err error) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.closedErr() != nil {
		return ss.closedErr()
	}

	ss.sessCtx.cancel(err)

	if err == nil {
		err = ErrInternalError
	}

	var annerr TerminateError
	if !errors.As(err, &annerr) {
		annerr = ErrInternalError
	}

	code := quic.StreamErrorCode(annerr.TerminateErrorCode())

	ss.stream.CancelRead(code)
	ss.stream.CancelWrite(code)

	slog.Debug("closed a session stream with an error",
		slog.Any("stream_id", ss.stream.StreamID()),
		slog.String("reason", err.Error()),
	)

	return nil
}

func (ss *sessionStream) listenUpdates() {
	var sum message.SessionUpdateMessage
	for {
		_, err := sum.Decode(ss.stream)
		if err != nil {
			slog.Error("failed to decode session update message", "error", err)
			return
		}

		slog.Debug("received a session update message",
			"stream_id", ss.stream.StreamID(),
			"bitrate", sum.Bitrate,
		)

		// TODO: Handle the session update message
	}
}

func (ss *sessionStream) closedErr() error {
	if ss.sessCtx.Err() != nil {
		reason := context.Cause(ss.sessCtx)
		if reason != nil {
			return reason
		}
		return ErrClosedSession
	}

	return nil
}

//
