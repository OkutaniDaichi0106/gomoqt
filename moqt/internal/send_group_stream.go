package internal

import (
	"errors"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newSendGroupStream(gm *message.GroupMessage, stream transport.SendStream) *SendGroupStream {
	return &SendGroupStream{
		GroupMessage: *gm,
		SendStream:   stream,
		startTime:    time.Now(),
	}
}

type SendGroupStream struct {
	GroupMessage message.GroupMessage
	SendStream   transport.SendStream
	startTime    time.Time
}

func (sgs *SendGroupStream) WriteFrameBytes(frame []byte) error {
	fm := message.FrameMessage{
		Payload: frame,
	}
	_, err := fm.Encode(sgs.SendStream)

	if err != nil {
		// Signal the group error code
		var grperr GroupError
		var code protocol.GroupErrorCode
		if errors.As(err, &grperr) {
			code = grperr.GroupErrorCode()
		} else {
			code = ErrInternalError.GroupErrorCode()
		}

		sgs.SendStream.CancelWrite(transport.StreamErrorCode(code))

		return err
	}

	return nil
}

func (sgs *SendGroupStream) StartAt() time.Time {
	return sgs.startTime
}

func (sgs *SendGroupStream) Close() error {
	return sgs.SendStream.Close()
}

func (sgs *SendGroupStream) SetWriteDeadline(t time.Time) error {
	return sgs.SendStream.SetWriteDeadline(t)
}

func (sgs *SendGroupStream) CancelWrite(code protocol.GroupErrorCode) {
	sgs.SendStream.CancelWrite(transport.StreamErrorCode(code))
}
