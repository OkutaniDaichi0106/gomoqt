package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
)

type ReceiveGroupStream interface {
	GroupReader

	SubscribeID() SubscribeID

	CancelRead(StreamErrorCode)

	SetReadDeadline(time.Time) error
}

var _ ReceiveGroupStream = (*receiveGroupStream)(nil)

type receiveGroupStream struct {
	internalStream *internal.ReceiveGroupStream
}

func (s *receiveGroupStream) GroupSequence() GroupSequence {
	return GroupSequence(s.internalStream.GroupMessage.GroupSequence)
}

func (s *receiveGroupStream) ReadFrame() ([]byte, error) {
	return s.internalStream.ReadFrame()
}

func (s *receiveGroupStream) CancelRead(code StreamErrorCode) {
	s.internalStream.CancelRead(internal.StreamErrorCode(code))
}

func (s *receiveGroupStream) SubscribeID() SubscribeID {
	return SubscribeID(s.internalStream.GroupMessage.SubscribeID)
}

func (s *receiveGroupStream) SetReadDeadline(t time.Time) error {
	return s.internalStream.SetReadDeadline(t)
}
