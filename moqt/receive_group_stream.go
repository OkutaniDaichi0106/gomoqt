package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

var _ GroupReader = (*receiveGroupStream)(nil)

type receiveGroupStream struct {
	internalStream *internal.ReceiveGroupStream
}

func (s *receiveGroupStream) GroupSequence() GroupSequence {
	return GroupSequence(s.internalStream.GroupMessage.GroupSequence)
}

func (s *receiveGroupStream) ReadFrame() (Frame, error) {
	bytes, err := s.internalStream.ReceiveFrameBytes()
	if err != nil {
		return nil, err
	}
	return NewFrame(bytes), nil
}

func (s *receiveGroupStream) CancelRead(err GroupError) {
	s.internalStream.CancelRead(protocol.GroupErrorCode(err.GroupErrorCode()))
}

func (s *receiveGroupStream) SetReadDeadline(t time.Time) error {
	return s.internalStream.SetReadDeadline(t)
}
