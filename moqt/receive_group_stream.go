package moqt

import (
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
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
	id       SubscribeID
	sequence GroupSequence
	stream   quic.ReceiveStream
}

func (s *receiveGroupStream) GroupSequence() GroupSequence {
	return s.sequence
}

func (s *receiveGroupStream) ReadFrame() (Frame, error) {
	var fm message.FrameMessage
	_, err := fm.Decode(s.stream)
	if err != nil {
		slog.Error("failed to decode a FRAME message", "error", err)
		return nil, err
	}

	slog.Info("received a FRAME message", slog.String("payload", string(fm.Payload)))

	return &fm, nil
}

func (s *receiveGroupStream) CancelRead(err GroupError) {
	s.stream.CancelRead(quic.StreamErrorCode(err.GroupErrorCode()))
}

func (s *receiveGroupStream) SetReadDeadline(t time.Time) error {
	err := s.stream.SetReadDeadline(t)
	if err != nil {
		slog.Error("failed to set read deadline",
			"error", err,
			"deadline", t.String(),
			"stream_id", s.stream.StreamID(),
			"subscribe_id", s.id,
			"sequence", s.sequence,
		)
		return err
	}

	slog.Info("set read deadline successfully",
		slog.String("deadline", t.String()),
	)

	return nil
}
