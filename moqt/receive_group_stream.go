package moqt

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

var _ GroupReader = (*receiveGroupStream)(nil)

func newReceiveGroupStream(trackCtx context.Context, sequence GroupSequence, stream quic.ReceiveStream) *receiveGroupStream {
	ctx, cancel := context.WithCancelCause(trackCtx)
	return &receiveGroupStream{
		sequence: sequence,
		stream:   stream,
		ctx:      ctx,
		cancel:   cancel,
	}
}

type receiveGroupStream struct {
	sequence GroupSequence

	stream     quic.ReceiveStream
	frameCount int64

	ctx    context.Context
	cancel context.CancelCauseFunc
}

func (s *receiveGroupStream) GroupSequence() GroupSequence {
	return s.sequence
}

func (s *receiveGroupStream) ReadFrame() (*Frame, error) {
	if err := s.ctx.Err(); err != nil {
		// If the context is already cancelled, return the error
		return nil, err
	}

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

			//
			s.cancel(grpErr)

			return nil, grpErr
		}

		return nil, err
	}
	s.frameCount++

	return frame, nil
}

func (s *receiveGroupStream) CancelRead(code GroupErrorCode) {
	if s.ctx.Err() != nil {
		// If the context is already cancelled, do nothing
		return
	}

	strErrCode := quic.StreamErrorCode(code)
	s.stream.CancelRead(strErrCode)

	grpErr := &GroupError{
		StreamError: &quic.StreamError{
			StreamID:  s.stream.StreamID(),
			ErrorCode: strErrCode,
		},
	}

	s.cancel(grpErr)
}

func (s *receiveGroupStream) SetReadDeadline(t time.Time) error {
	return s.stream.SetReadDeadline(t)
}
