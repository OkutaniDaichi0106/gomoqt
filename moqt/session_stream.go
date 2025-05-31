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

func newSessionStream(sessCtx *sessionContext, stream quic.Stream, tracer *moqtrace.StreamTracer) *sessionStream {
	sess := &sessionStream{
		updatedCh: make(chan struct{}, 1),
		stream:    stream,
		tracer:    tracer,
	}

	sess.listenOnce.Do(func() {
		sessCtx.wg.Add(1)
		go func() {
			defer sessCtx.wg.Done()
			sess.listenUpdates(sessCtx)
		}()
	})

	go func() {
		<-sessCtx.Done()
		if sessCtx.Err() != nil {
			reason := context.Cause(sessCtx)
			sess.closeWithError(reason)
			return
		}

		sess.close()
	}()

	return sess
}

type sessionStream struct {
	updatedCh chan struct{}

	localBitrate  uint64 // The bitrate set by the local
	remoteBitrate uint64 // The bitrate set by the remote

	stream quic.Stream
	mu     sync.Mutex

	listenOnce sync.Once

	closed   bool
	closeErr error

	tracer *moqtrace.StreamTracer // Tracer for the stream
}

func (ss *sessionStream) updateSession(bitrate uint64) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	sum := message.SessionUpdateMessage{
		Bitrate: bitrate,
	}
	_, err := sum.Encode(ss.stream)
	if err != nil {
		slog.Error("failed to send a SESSION_UPDATE message", "error", err)
		return err
	}
	ss.tracer.SessionUpdateMessageSent(sum)

	ss.localBitrate = bitrate

	return nil
}

func (ss *sessionStream) SessionUpdated() <-chan struct{} {
	return ss.updatedCh
}

func (ss *sessionStream) close() error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.closed {
		return ss.closeErr
	}

	ss.closed = true

	return ss.stream.Close()
}

func (ss *sessionStream) closeWithError(reason error) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if reason == nil {
		reason = ErrInternalError
	}

	if ss.closed {
		if ss.closeErr != nil {
			return ss.closeErr
		}
		return nil
	}

	ss.closed = true

	var trmerr TerminateError
	if !errors.As(reason, &trmerr) {
		trmerr = ErrInternalError.WithReason(reason.Error())
	}

	code := quic.StreamErrorCode(trmerr.TerminateErrorCode())

	ss.stream.CancelRead(code)
	ss.stream.CancelWrite(code)

	ss.tracer.ReceiveStreamCancelled(code, trmerr.Error())
	ss.tracer.SendStreamCancelled(code, trmerr.Error())

	return nil
}

func (ss *sessionStream) listenUpdates(sessCtx *sessionContext) {
	var sum message.SessionUpdateMessage
	var err error

	logger := sessCtx.Logger()

	for {
		if ss.closed {
			return
		}

		_, err = sum.Decode(ss.stream)
		if err != nil {
			if logger != nil {
				logger.Error("failed to decode session update message",
					"error", err,
				)
			}

			sessCtx.cancel(err)

			return
		}
		ss.tracer.SessionUpdateMessageReceived(sum)

		// Update the session bitrate
		ss.mu.Lock()
		ss.remoteBitrate = sum.Bitrate
		ss.mu.Unlock()

		// Notify that the session has been updated
		select {
		case ss.updatedCh <- struct{}{}:
		default:
		}
	}
}
