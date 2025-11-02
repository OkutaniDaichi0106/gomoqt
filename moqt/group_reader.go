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
		// frame:    newFrame(0),
		onClose: onClose,
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
		panic("nil frame")
	}
	err := frame.decode(s.stream)
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

func (s *GroupReader) Frames(buf *Frame) func(yield func(*Frame) bool) {
	return func(yield func(*Frame) bool) {
		if buf == nil {
			buf = NewFrame(0)
		}
		var err error
		for {
			err = s.ReadFrame(buf)
			if err != nil {
				return
			}

			if !yield(buf) {
				return
			}
		}
	}
}
