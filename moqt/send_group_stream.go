package moqt

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

var _ GroupWriter = (*sendGroupStream)(nil)

type sendGroupStream struct {
	internalStream *internal.SendGroupStream

	closed    bool
	closedErr error
	mu        sync.Mutex
}

func (sgs *sendGroupStream) GroupSequence() GroupSequence {
	return GroupSequence(sgs.internalStream.GroupMessage.GroupSequence)
}

func (sgs *sendGroupStream) WriteFrame(frame *Frame) error {
	sgs.mu.Lock()
	defer sgs.mu.Unlock()

	if sgs.closed {
		if sgs.closedErr != nil {
			return sgs.closedErr
		}

		return ErrClosedGroup
	}

	err := sgs.internalStream.WriteFrameBytes(frame.bytes)
	if err != nil {
		sgs.CloseWithError(err) // TODO: should we close the stream?
		return err
	}

	return nil
}

func (sgs *sendGroupStream) SetWriteDeadline(t time.Time) error {
	return sgs.internalStream.SetWriteDeadline(t)
}

func (sgs *sendGroupStream) CloseWithError(err error) error {
	sgs.mu.Lock()
	defer sgs.mu.Unlock()

	if sgs.closed {
		if sgs.closedErr != nil {
			return fmt.Errorf("stream has already closed due to: %w", sgs.closedErr)
		}
		return errors.New("stream has already closed")
	}

	if err == nil {
		err = ErrInternalError
	}

	var grperr GroupError
	if !errors.As(err, &grperr) {
		errors.As(ErrInternalError, &grperr)
	}

	sgs.internalStream.CancelWrite(protocol.GroupErrorCode(grperr.GroupErrorCode()))

	sgs.closed = true
	sgs.closedErr = err

	return nil
}

func (sgs *sendGroupStream) Close() error {
	sgs.mu.Lock()
	defer sgs.mu.Unlock()

	if sgs.closed {
		if sgs.closedErr != nil {
			return fmt.Errorf("stream has already closed due to: %w", sgs.closedErr)
		}
		return errors.New("stream has already closed")
	}

	sgs.closed = true

	return sgs.internalStream.Close()
}
