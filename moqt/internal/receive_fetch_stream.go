package internal

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newReceiveFetchStream(fm *message.FetchMessage, stream transport.Stream) *ReceiveFetchStream {
	return &ReceiveFetchStream{
		FetchMessage: *fm,
		Stream:       stream,
	}
}

type ReceiveFetchStream struct {
	FetchMessage message.FetchMessage
	Stream       transport.Stream
	mu           sync.Mutex
	closed       bool
	closeErr     error
}

func (rfs *ReceiveFetchStream) CancelWrite(code message.GroupErrorCode) {
	rfs.Stream.CancelWrite(transport.StreamErrorCode(code))
}

func (rfs *ReceiveFetchStream) SetWriteDeadline(t time.Time) error {
	return rfs.Stream.SetWriteDeadline(t)
}

func (rfs *ReceiveFetchStream) WriteFrame(frame []byte) error {
	fm := message.FrameMessage{
		Payload: frame,
	}
	_, err := fm.Encode(rfs.Stream)
	if err != nil {
		return err
	}

	return nil
}

func (rfs *ReceiveFetchStream) Close() error {
	return rfs.Stream.Close()
}

func (rfs *ReceiveFetchStream) CloseWithError(err error) error {
	rfs.mu.Lock()
	defer rfs.mu.Unlock()

	slog.Debug("closing a receive fetch stream with an error", slog.String("error", err.Error()))

	if rfs.closed {
		return rfs.closeErr
	}

	if err == nil {
		return rfs.Close()
	}

	rfs.closeErr = err
	rfs.closed = true

	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		feterr, ok := err.(FetchError)
		if ok {
			code = transport.StreamErrorCode(feterr.FetchErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	rfs.Stream.CancelRead(code)
	rfs.Stream.CancelWrite(code)

	slog.Debug("closed a receive fetch stream with an error", slog.String("error", err.Error()))

	return nil
}
