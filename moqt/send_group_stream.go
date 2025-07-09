package moqt

import (
	"context"
	"errors"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

var _ GroupWriter = (*sendGroupStream)(nil)

func newSendGroupStream(trackCtx context.Context, stream quic.SendStream, sequence GroupSequence) *sendGroupStream {
	ctx, cancel := context.WithCancelCause(trackCtx)
	return &sendGroupStream{
		ctx:      ctx,
		cancel:   cancel,
		sequence: sequence,
		stream:   stream,
	}
}

type sendGroupStream struct {
	ctx    context.Context
	cancel context.CancelCauseFunc

	sequence GroupSequence

	stream quic.SendStream

	frameCount uint64 // Number of frames sent on this stream
}

func (sgs *sendGroupStream) GroupSequence() GroupSequence {
	return sgs.sequence
}

func (sgs *sendGroupStream) WriteFrame(frame *Frame) error {
	if frame == nil || frame.message == nil {
		return errors.New("frame is nil or has no bytes")
	}

	if err := sgs.ctx.Err(); err != nil {
		// If the context is already cancelled, return the error
		return err
	}

	err := frame.message.Encode(sgs.stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			grpErr := &GroupError{
				StreamError: strErr,
			}

			sgs.cancel(grpErr)

			return grpErr
		}

		sgs.cancel(err)

		return err
	}

	sgs.frameCount++

	return nil
}

func (sgs *sendGroupStream) SetWriteDeadline(t time.Time) error {
	return sgs.stream.SetWriteDeadline(t)
}

func (sgs *sendGroupStream) CancelWrite(code GroupErrorCode) error {
	if err := sgs.ctx.Err(); err != nil {
		return err
	}

	strErrCode := quic.StreamErrorCode(code)
	sgs.stream.CancelWrite(strErrCode)

	grpErr := &GroupError{
		StreamError: &quic.StreamError{
			StreamID:  sgs.stream.StreamID(),
			ErrorCode: strErrCode,
		},
	}

	sgs.cancel(grpErr)

	return nil
}

func (sgs *sendGroupStream) Close() error {
	if err := sgs.ctx.Err(); err != nil {
		return err
	}

	err := sgs.stream.Close()
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			grpErr := &GroupError{
				StreamError: strErr,
			}
			sgs.cancel(grpErr)

			return grpErr
		}
		return err
	}

	// Successfully closed the stream, cancel the context
	sgs.cancel(nil)

	return nil
}
