package moqt

import (
	"errors"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

var _ GroupWriter = (*sendGroupStream)(nil)

type sendGroupStream struct {
	internalStream *internal.SendGroupStream
	// groupErrCh     chan GroupErrorCode
}

func (sgs *sendGroupStream) GroupSequence() GroupSequence {
	return GroupSequence(sgs.internalStream.GroupMessage.GroupSequence)
}

func (sgs *sendGroupStream) CancelWrite(code GroupErrorCode) {
	sgs.internalStream.SendStream.CancelWrite(transport.StreamErrorCode(code))
}

func (sgs *sendGroupStream) SetWriteDeadline(t time.Time) error {
	return sgs.internalStream.SetWriteDeadline(t)
}

func (sgs *sendGroupStream) Close() error {
	return sgs.internalStream.Close()
}

func (sgs *sendGroupStream) WriteFrame(frame []byte) error {
	err := sgs.internalStream.WriteFrame(frame)
	if err != nil {
		var grperr GroupError
		if errors.As(err, &grperr) {
			sgs.CancelWrite(GroupErrorCode(grperr.GroupErrorCode()))
		}

		return err
	}

	return nil
}

// methods for relaying bytes
var _ directBytesWriter = (*sendGroupStream)(nil)

func (sgs *sendGroupStream) newBytesWriter() writer {
	return &streamBytesWriter{sgs}
}

var _ writer = (*streamBytesWriter)(nil)

type streamBytesWriter struct {
	stream *sendGroupStream
}

func (s *streamBytesWriter) Write(p *[]byte) (int, error) {
	return s.stream.internalStream.SendStream.Write(*p)
}
