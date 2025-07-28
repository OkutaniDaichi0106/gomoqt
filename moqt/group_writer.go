package moqt

import (
	"context"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newGroupWriter(stream quic.SendStream, sequence GroupSequence,
	onClose func()) *GroupWriter {

	return &GroupWriter{
		sequence: sequence,
		onClose:  onClose,
		stream:   stream,
		ctx:      context.WithValue(stream.Context(), &uniStreamTypeCtxKey, message.StreamTypeGroup),
	}
}

type GroupWriter struct {
	sequence GroupSequence

	ctx    context.Context
	stream quic.SendStream

	frameCount uint64 // Number of frames sent on this stream

	onClose func()
}

func (sgs *GroupWriter) GroupSequence() GroupSequence {
	return sgs.sequence
}

func (sgs *GroupWriter) WriteFrame(frame *Frame) error {
	if frame == nil || frame.message == nil {
		return nil
	}

	err := frame.message.Encode(sgs.stream)
	if err != nil {
		return Cause(sgs.ctx)
	}

	sgs.frameCount++

	return nil
}

func (sgs *GroupWriter) SetWriteDeadline(t time.Time) error {
	return sgs.stream.SetWriteDeadline(t)
}

func (sgs *GroupWriter) CancelWrite(code GroupErrorCode) {
	sgs.stream.CancelWrite(quic.StreamErrorCode(code))

	sgs.onClose()
}

func (sgs *GroupWriter) Close() error {
	err := sgs.stream.Close()
	if err != nil {
		return Cause(sgs.ctx)
	}

	sgs.onClose()

	return nil
}

func (s *GroupWriter) Context() context.Context {
	return s.ctx
}
