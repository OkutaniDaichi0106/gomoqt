package moqt

import (
	"errors"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

var _ GroupWriter = (*sendGroupStream)(nil)

func newSendGroupStream(stream quic.SendStream, id SubscribeID, sequence GroupSequence) *sendGroupStream {
	return &sendGroupStream{
		id:       id,
		sequence: sequence,
		stream:   stream,
	}
}

type sendGroupStream struct {
	id       SubscribeID
	sequence GroupSequence
	stream   quic.SendStream

	closed    bool
	closedErr error
	mu        sync.Mutex
}

func (sgs *sendGroupStream) GroupSequence() GroupSequence {
	return sgs.sequence
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

	if sgs.closed {
		if sgs.closedErr != nil {
			return sgs.closedErr
		}
		return nil
	}

	if err == nil {
		err = ErrInternalError
	}

	var grperr GroupError
	if !errors.As(err, &grperr) {
		errors.As(ErrInternalError, &grperr)
	}

	sgs.stream.CancelWrite(quic.StreamErrorCode(grperr.GroupErrorCode()))

	sgs.closed = true
	sgs.closedErr = err

	return nil
}

func (sgs *sendGroupStream) Close() error {
	sgs.mu.Lock()
	defer sgs.mu.Unlock()

	if sgs.closed {
		if sgs.closedErr != nil {
			return sgs.closedErr
		}
		return nil
	}

	sgs.closed = true

	return sgs.stream.Close()
}
