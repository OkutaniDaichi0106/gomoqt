package moqt

import (
	"errors"
	"io"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

var _ GroupReader = (*receiveGroupStream)(nil)

func newReceiveGroupStream(sequence GroupSequence, stream quic.ReceiveStream) *receiveGroupStream {
	return &receiveGroupStream{
		sequence: sequence,
		stream:   stream,
		doneCh:   make(chan struct{}),
	}

}

type receiveGroupStream struct {
	sequence GroupSequence

	stream     quic.ReceiveStream
	frameCount int64

	doneCh    chan struct{}
	readErr   error
	cancelled bool
}

func (s *receiveGroupStream) GroupSequence() GroupSequence {
	return s.sequence
}

func (s *receiveGroupStream) ReadFrame() (*Frame, error) {
	frame := NewFrame(nil)
	_, err := frame.message.Decode(s.stream)
	if err != nil {
		if !s.cancelled {
			s.cancelled = true
			close(s.doneCh)
		}

		if errors.Is(err, io.EOF) {
			return nil, err
		}

		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			// Stream was reset
			s.readErr = &GroupError{
				StreamError: strErr,
			}

			return nil, &GroupError{
				StreamError: strErr,
			}
		}

		return nil, err
	}
	s.frameCount++

	return frame, nil
}

func (s *receiveGroupStream) CancelRead(code GroupErrorCode) {
	strErrCode := quic.StreamErrorCode(code)
	s.stream.CancelRead(strErrCode)

	if !s.cancelled {
		s.cancelled = true

		s.readErr = &GroupError{
			StreamError: &quic.StreamError{
				StreamID:  s.stream.StreamID(),
				ErrorCode: quic.StreamErrorCode(code),
			},
		}

		close(s.doneCh)
	}
}

func (s *receiveGroupStream) SetReadDeadline(t time.Time) error {
	return s.stream.SetReadDeadline(t)
}
