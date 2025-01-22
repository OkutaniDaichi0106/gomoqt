package moqt

import (
	"errors"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

/*
 * Group Writer
 */
type GroupWriter interface {
	GroupSequence() GroupSequence
	WriteFrame([]byte) error
	Close() error
}

/*
 * Group Sender
 */
type SendGroupStream interface {
	GroupWriter

	SubscribeID() SubscribeID

	CancelWrite(GroupErrorCode)
	SetWriteDeadline(time.Time) error
}

var _ SendGroupStream = (*sendGroupStream)(nil)

type sendGroupStream struct {
	sequence    GroupSequence
	stream      transport.SendStream
	subscribeID SubscribeID
	startTime   time.Time

	errCodeCh chan GroupErrorCode
}

func (stream sendGroupStream) SubscribeID() SubscribeID {
	return stream.subscribeID
}

func (stream sendGroupStream) GroupSequence() GroupSequence {
	return stream.sequence
}

func (stream sendGroupStream) WriteFrame(buf []byte) error {
	fm := message.FrameMessage{
		Payload: buf,
	}
	err := fm.Encode(stream.stream)

	if err != nil {
		// Signal the group error code
		var strerr transport.StreamError
		var code GroupErrorCode
		if errors.As(err, &strerr) {
			code = GroupErrorCode(strerr.StreamErrorCode())
		} else {
			code = ErrInternalError.GroupErrorCode()
		}

		stream.CancelWrite(code)

		return err
	}

	return nil
}

func (stream sendGroupStream) StartAt() time.Time {
	return stream.startTime
}

func (stream sendGroupStream) Close() error {
	return stream.stream.Close()
}

func (stream sendGroupStream) CancelWrite(code GroupErrorCode) {
	if stream.errCodeCh == nil {
		stream.errCodeCh = make(chan GroupErrorCode, 1)
	}

	select {
	case stream.errCodeCh <- code:
	default:
	}

	stream.stream.CancelWrite(transport.StreamErrorCode(code))
}

func (stream sendGroupStream) SetWriteDeadline(t time.Time) error {
	err := stream.stream.SetWriteDeadline(t)

	// Signal the group error code
	if err != nil {
		// Signal the group error code
		var strerr transport.StreamError
		var code GroupErrorCode
		if errors.As(err, &strerr) {
			code = GroupErrorCode(strerr.StreamErrorCode())
		} else {
			code = ErrInternalError.GroupErrorCode()
		}

		stream.CancelWrite(code)

		return err
	}

	return nil
}
