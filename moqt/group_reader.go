package moqt

import (
	"errors"
	"io"
	"time"

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

func (s *GroupReader) ReadFrame() (*Frame, error) {
	frame := NewFrame(nil)
	err := frame.message.Decode(s.stream)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, err
		}

		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			grpErr := &GroupError{
				StreamError: strErr,
			}

			return nil, grpErr
		}

		return nil, err
	}
	s.frameCount++

	return frame, nil
}

func (s *GroupReader) CancelRead(code GroupErrorCode) {
	strErrCode := quic.StreamErrorCode(code)
	s.stream.CancelRead(strErrCode)
}

func (s *GroupReader) SetReadDeadline(t time.Time) error {
	return s.stream.SetReadDeadline(t)
}
