package moqt

import (
	"errors"
	"io"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

func newGroupReader(sequence GroupSequence, stream quic.ReceiveStream,
	onClose func()) *GroupReader {
	return &GroupReader{
		sequence: sequence,
		stream:   stream,
		frame:    newFrame(0),
		onClose:  onClose,
	}
}

type GroupReader struct {
	sequence GroupSequence

	stream     quic.ReceiveStream
	frameCount int64

	frame *Frame

	onClose func()
}

func (s *GroupReader) GroupSequence() GroupSequence {
	return s.sequence
}

func (s *GroupReader) ReadFrame() (*Frame, error) {
	if s.frame == nil {
		s.frame = newFrame(0)
	}
	err := s.frame.decode(s.stream)
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

	return s.frame, nil
}

func (s *GroupReader) CancelRead(code GroupErrorCode) {
	strErrCode := quic.StreamErrorCode(code)
	s.stream.CancelRead(strErrCode)
}

func (s *GroupReader) SetReadDeadline(t time.Time) error {
	return s.stream.SetReadDeadline(t)
}

func (s *GroupReader) Frames() func(yield func(*Frame) bool) {
	return func(yield func(*Frame) bool) {
		for {
			frame, err := s.ReadFrame()
			if err != nil {
				return
			}

			if !yield(frame) {
				return
			}
		}
	}
}
