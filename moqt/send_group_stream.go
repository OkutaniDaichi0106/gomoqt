package moqt

import (
	"context"
	"errors"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newSendGroupStream(stream quic.SendStream, sequence GroupSequence,
	onClose func()) *GroupWriter {

	return &GroupWriter{
		sequence: sequence,
		onClose:  onClose,
		stream:   stream,
	}
}

type GroupWriter struct {
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

	if ctx := sgs.stream.Context(); ctx.Err() != nil {
		// If the context is already cancelled, return the error
		reason := context.Cause(ctx)
		var strErr *quic.StreamError
		if errors.As(reason, &strErr) {
			return &GroupError{
				StreamError: strErr,
			}
		}

		return reason
	}

	err := frame.message.Encode(sgs.stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			grpErr := &GroupError{
				StreamError: strErr,
			}

			return grpErr
		}

		return err
	}

	sgs.frameCount++

	return nil
}

func (sgs *GroupWriter) SetWriteDeadline(t time.Time) error {
	return sgs.stream.SetWriteDeadline(t)
}

func (sgs *GroupWriter) CancelWrite(code GroupErrorCode) {
	strErrCode := quic.StreamErrorCode(code)
	sgs.stream.CancelWrite(strErrCode)

	sgs.onClose()
}

func (sgs *GroupWriter) Close() error {
	err := sgs.stream.Close()
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			grpErr := &GroupError{
				StreamError: strErr,
			}

			return grpErr
		}
		return err
	}

	sgs.onClose()

	return nil
}
