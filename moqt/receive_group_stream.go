package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

var _ GroupReader = (*receiveGroupStream)(nil)

type receiveGroupStream struct {
	internalStream *internal.ReceiveGroupStream
}

func (s *receiveGroupStream) GroupSequence() GroupSequence {
	return GroupSequence(s.internalStream.GroupMessage.GroupSequence)
}

func (s *receiveGroupStream) ReadFrame() ([]byte, error) {
	return s.internalStream.ReadFrame()
}

func (s *receiveGroupStream) CancelRead(code GroupErrorCode) {
	s.internalStream.CancelRead(protocol.GroupErrorCode(code))
}

func (s *receiveGroupStream) SetReadDeadline(t time.Time) error {
	return s.internalStream.SetReadDeadline(t)
}

// methods for relaying bytes
var _ directBytesReader = (*receiveGroupStream)(nil)

func (s *receiveGroupStream) newBytesReader() reader {
	return &streamBytesReader{s.internalStream.ReceiveStream}
}

var _ reader = (*streamBytesReader)(nil)

type streamBytesReader struct {
	stream transport.ReceiveStream
}

func (s *streamBytesReader) Read(p *[]byte) (int, error) {
	return s.stream.Read(*p)
}
