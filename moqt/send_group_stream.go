package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

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
	internalStream *internal.SendGroupStream
}

func (sgs *sendGroupStream) SubscribeID() SubscribeID {
	return SubscribeID(sgs.internalStream.GroupMessage.SubscribeID)
}

func (sgs *sendGroupStream) GroupSequence() GroupSequence {
	return GroupSequence(sgs.internalStream.GroupMessage.GroupSequence)
}

func (sgs *sendGroupStream) CancelWrite(code GroupErrorCode) {
	sgs.internalStream.CancelWrite(message.GroupErrorCode(code))
}

func (sgs *sendGroupStream) SetWriteDeadline(t time.Time) error {
	return sgs.internalStream.SetWriteDeadline(t)
}

func (sgs *sendGroupStream) Close() error {
	return sgs.internalStream.Close()
}

func (sgs *sendGroupStream) WriteFrame(frame []byte) error {
	return sgs.internalStream.WriteFrame(frame)
}
