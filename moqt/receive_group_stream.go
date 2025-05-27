package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

var _ GroupReader = (*receiveGroupStream)(nil)

func newReceiveGroupStream(id SubscribeID, sequence GroupSequence, stream quic.ReceiveStream) *receiveGroupStream {
	return &receiveGroupStream{
		id:       id,
		sequence: sequence,
		stream:   stream,
	}
}

type receiveGroupStream struct {
	groupCtx *groupContext
	id       SubscribeID
	sequence GroupSequence
	stream   quic.ReceiveStream
}

func (s *receiveGroupStream) GroupSequence() GroupSequence {
	return s.sequence
}

func (s *receiveGroupStream) ReadFrame() (*Frame, error) {
	frame := NewFrame(nil)
	_, err := frame.message.Decode(s.stream)
	if err != nil {
		if logger := s.groupCtx.Logger(); logger != nil {
			logger.Error("failed to decode a FRAME message", "error", err)
		}
		return nil, err
	}

	return frame, nil
}

func (s *receiveGroupStream) CancelRead(err GroupError) {
	s.stream.CancelRead(quic.StreamErrorCode(err.GroupErrorCode()))
}

func (s *receiveGroupStream) SetReadDeadline(t time.Time) error {
	err := s.stream.SetReadDeadline(t)
	if err != nil {
		if logger := s.groupCtx.Logger(); logger != nil {
			logger.Error("failed to set read deadline",
				"error", err,
			)
		}
		return err
	}

	return nil
}
