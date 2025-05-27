package moqt

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

var _ GroupWriter = (*sendGroupStream)(nil)

func newSendGroupStream(stream quic.SendStream, groupCtx *groupContext) *sendGroupStream {
	return &sendGroupStream{
		groupCtx: groupCtx,
		stream:   stream,
	}
}

type sendGroupStream struct {
	groupCtx *groupContext

	stream quic.SendStream

	mu sync.Mutex
}

func (sgs *sendGroupStream) GroupSequence() GroupSequence {
	return sgs.groupCtx.seq
}

func (sgs *sendGroupStream) WriteFrame(frame *Frame) error {
	sgs.mu.Lock()
	defer sgs.mu.Unlock()

	if err := sgs.closedErr(); err != nil {
		return err
	}

	if frame == nil {
		return errors.New("frame is nil")
	}

	_, err := frame.message.Encode(sgs.stream)
	if err != nil {
		return err
	}

	// TODO: Consider waiting briefly before sending if the frame queue is full.

	return nil
}

func (sgs *sendGroupStream) SetWriteDeadline(t time.Time) error {
	return sgs.stream.SetWriteDeadline(t)
}

func (sgs *sendGroupStream) CloseWithError(err error) error {
	sgs.mu.Lock()
	defer sgs.mu.Unlock()

	if err := sgs.closedErr(); err != nil {
		return err
	}

	sgs.groupCtx.cancel(err)

	if err == nil {
		err = ErrInternalError
	}

	var grperr GroupError
	if !errors.As(err, &grperr) {
		grperr = ErrInternalError
	}

	sgs.stream.CancelWrite(quic.StreamErrorCode(grperr.GroupErrorCode()))

	return nil
}

func (sgs *sendGroupStream) Close() error {
	sgs.mu.Lock()
	defer sgs.mu.Unlock()

	if err := sgs.closedErr(); err != nil {
		return err
	}

	sgs.groupCtx.cancel(ErrClosedGroup)

	err := sgs.stream.Close()
	if err != nil {
		return err
	}

	return sgs.stream.Close()
}

func (sgs *sendGroupStream) closedErr() error {
	if sgs.groupCtx.Err() != nil {
		reason := context.Cause(sgs.groupCtx)
		if reason != nil {
			return reason
		}
		return ErrClosedGroup
	}

	return nil
}
