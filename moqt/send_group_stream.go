package moqt

import (
	"context"
	"errors"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSendGroupStream(trackCtx context.Context, stream quic.SendStream, sequence GroupSequence,
	onClose func()) *GroupWriter {
	ctx, cancel := context.WithCancelCause(trackCtx)
	go func() {
		streamCtx := stream.Context()
		<-streamCtx.Done()
		reason := context.Cause(streamCtx)
		var (
			strErr *quic.StreamError
			appErr *quic.ApplicationError
		)
		if errors.As(reason, &strErr) {
			reason = &GroupError{
				StreamError: strErr,
			}
		} else if errors.As(reason, &appErr) {
			reason = &SessionError{
				ApplicationError: appErr,
			}
		}
		cancel(reason)
	}()
	return &GroupWriter{
		ctx:      ctx,
		cancel:   cancel,
		sequence: sequence,
		onClose:  onClose,
		stream:   stream,
	}
}

type GroupWriter struct {
	ctx    context.Context
	cancel context.CancelCauseFunc

	sequence GroupSequence

	stream quic.SendStream

	frameCount uint64 // Number of frames sent on this stream

	onClose func()
}

func (sgs *GroupWriter) GroupSequence() GroupSequence {
	return sgs.sequence
}

func (sgs *GroupWriter) WriteFrame(frame *Frame) error {
	if frame == nil || frame.message == nil {
		return errors.New("frame is nil or has no bytes")
	}

	if sgs.ctx.Err() != nil {
		// If the context is already cancelled, return the error
		return context.Cause(sgs.ctx)
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

func (sgs *GroupWriter) SetWriteDeadline(t time.Time) error {
	return sgs.stream.SetWriteDeadline(t)
}

func (sgs *GroupWriter) CancelWrite(code GroupErrorCode) {
	if sgs.ctx.Err() != nil {
		// If the context is already cancelled, local cancelled the stream or the remote side closed the stream
		// so we don't need to cancel it again
		return
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

	sgs.onClose()
}

func (sgs *GroupWriter) Close() error {
	if sgs.ctx.Err() != nil {
		return context.Cause(sgs.ctx)
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

	sgs.onClose()

	return nil
}
