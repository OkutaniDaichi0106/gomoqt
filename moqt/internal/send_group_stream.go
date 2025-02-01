package internal

import (
	"errors"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newSendGroupStream(gm *message.GroupMessage, stream transport.SendStream) *SendGroupStream {
	return &SendGroupStream{
		GroupMessage: *gm,
		Stream:       stream,
		startTime:    time.Now(),
		errCodeCh:    make(chan message.GroupErrorCode, 1),
	}
}

type SendGroupStream struct {
	GroupMessage message.GroupMessage
	Stream       transport.SendStream
	startTime    time.Time

	errCodeCh chan message.GroupErrorCode
}

func (sgs *SendGroupStream) WriteFrame(frame []byte) error {
	fm := message.FrameMessage{
		Payload: frame,
	}
	_, err := fm.Encode(sgs.Stream)

	if err != nil {
		// Signal the group error code
		var strerr transport.StreamError
		var code message.GroupErrorCode
		if errors.As(err, &strerr) {
			code = message.GroupErrorCode(strerr.StreamErrorCode())
		} else {
			code = ErrInternalError.GroupErrorCode()
		}

		sgs.CancelWrite(code)

		return err
	}

	return nil
}

func (sgs *SendGroupStream) StartAt() time.Time {
	return sgs.startTime
}

func (sgs *SendGroupStream) Close() error {
	return sgs.Stream.Close()
}

func (sgs *SendGroupStream) CancelWrite(code message.GroupErrorCode) {
	if sgs.errCodeCh == nil {
		sgs.errCodeCh = make(chan message.GroupErrorCode, 1)
	}

	select {
	case sgs.errCodeCh <- code:
	default:
	}

	sgs.Stream.CancelWrite(transport.StreamErrorCode(code))
}

func (sgs *SendGroupStream) SetWriteDeadline(t time.Time) error {
	err := sgs.Stream.SetWriteDeadline(t)

	// Signal the group error code
	if err != nil {
		// Signal the group error code
		var strerr transport.StreamError
		var code message.GroupErrorCode
		if errors.As(err, &strerr) {
			code = message.GroupErrorCode(strerr.StreamErrorCode())
		} else {
			code = ErrInternalError.GroupErrorCode()
		}

		sgs.CancelWrite(code)

		return err
	}

	return nil
}
