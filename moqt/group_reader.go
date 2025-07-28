package moqt

import (
	"errors"
	"io"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newGroupReader(sequence GroupSequence, stream quic.ReceiveStream,
	onClose func()) *GroupReader {
	return &GroupReader{
		sequence: sequence,
		stream:   stream,
		onClose:  onClose,
	}
}

type GroupReader struct {
	sequence GroupSequence

	stream     quic.ReceiveStream
	frameCount int64

	onClose func()
}

func (s *GroupReader) GroupSequence() GroupSequence {
	return s.sequence
}

func (s *GroupReader) ReadFrame(frame *Frame) error {
	if frame == nil {
		return errors.New("frame cannot be nil")
	}

	// Set the internal message if not already set
	if frame.message == nil {
		frame.message = &message.FrameMessage{}
	}

	err := frame.message.Decode(s.stream)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return err
		}

		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			grpErr := &GroupError{
				StreamError: strErr,
			}

			return grpErr
		}

		return err
	}

	s.frameCount++

	return nil
}

func (s *GroupReader) CancelRead(code GroupErrorCode) {
	strErrCode := quic.StreamErrorCode(code)
	s.stream.CancelRead(strErrCode)
}

func (s *GroupReader) SetReadDeadline(t time.Time) error {
	return s.stream.SetReadDeadline(t)
}
