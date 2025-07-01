package moqt

import (
	"errors"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

var _ GroupWriter = (*sendGroupStream)(nil)

func newSendGroupStream(stream quic.SendStream, sequence GroupSequence) *sendGroupStream {
	return &sendGroupStream{
		closedCh: make(chan struct{}, 1),
		sequence: sequence,
		stream:   stream,
	}
}

type sendGroupStream struct {
	closed   bool
	closeErr error
	closedCh chan struct{}

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

	err := frame.message.Encode(sgs.stream)
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			return &GroupError{
				StreamError: strErr,
			}
		}

		return err
	}

	sgs.frameCount++

	return nil
}

func (sgs *sendGroupStream) SetWriteDeadline(t time.Time) error {
	return sgs.stream.SetWriteDeadline(t)
}

func (sgs *sendGroupStream) CancelWrite(code GroupErrorCode) error {
	if sgs.closed {
		return sgs.closeErr
	}

	sgs.closed = true

	defer close(sgs.closedCh)

	strErrCode := quic.StreamErrorCode(code)
	sgs.stream.CancelWrite(strErrCode)

	err := &GroupError{
		StreamError: &quic.StreamError{
			StreamID:  sgs.stream.StreamID(),
			ErrorCode: strErrCode,
		},
	}

	sgs.closeErr = err

	return nil
}

func (sgs *sendGroupStream) Close() error {
	if sgs.closed {
		return sgs.closeErr
	}

	sgs.closed = true

	defer close(sgs.closedCh)

	err := sgs.stream.Close()
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			sgs.closeErr = &GroupError{
				StreamError: strErr,
			}
			return sgs.closeErr
		}
		return err
	}

	return nil
}
