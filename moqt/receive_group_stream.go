package moqt

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newReceiveGroupStream(trackCtx context.Context, sequence GroupSequence, stream quic.ReceiveStream) *GroupReader {
	ctx, cancel := context.WithCancelCause(trackCtx)
	return &GroupReader{
		sequence: sequence,
		stream:   stream,
		ctx:      ctx,
		cancel:   cancel,
	}
}

type GroupReader struct {
	sequence GroupSequence

	stream     quic.ReceiveStream
	frameCount int64

	ctx    context.Context
	cancel context.CancelCauseFunc
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

			//
			s.cancel(grpErr)

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

	grpErr := &GroupError{
		StreamError: &quic.StreamError{
			StreamID:  s.stream.StreamID(),
			ErrorCode: strErrCode,
		},
	}

	s.cancel(grpErr)
}

func (s *GroupReader) SetReadDeadline(t time.Time) error {
	return s.stream.SetReadDeadline(t)
}
