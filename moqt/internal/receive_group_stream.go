package internal

import (
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newReceiveGroupStream(gm *message.GroupMessage, stream transport.ReceiveStream) *ReceiveGroupStream {
	return &ReceiveGroupStream{
		GroupMessage:  *gm,
		ReceiveStream: stream,
		startTime:     time.Now(),
	}
}

type ReceiveGroupStream struct {
	GroupMessage  message.GroupMessage
	ReceiveStream transport.ReceiveStream

	startTime time.Time
}

func (r ReceiveGroupStream) ReceiveFrameBytes() ([]byte, error) {
	var fm message.FrameMessage
	_, err := fm.Decode(r.ReceiveStream)
	if err != nil {
		slog.Error("failed to decode a FRAME message", "error", err)
		return nil, err
	}

	slog.Info("received a FRAME message", slog.String("payload", string(fm.Payload)))

	return fm.Payload, nil
}

func (r ReceiveGroupStream) StartAt() time.Time {
	return r.startTime
}

func (r ReceiveGroupStream) CancelRead(code protocol.GroupErrorCode) {
	r.ReceiveStream.CancelRead(transport.StreamErrorCode(code))

	slog.Debug("canceled read", slog.Any("code", code))
}

func (r ReceiveGroupStream) SetReadDeadline(t time.Time) error {
	err := r.ReceiveStream.SetReadDeadline(t)
	if err != nil {
		slog.Error("failed to set read deadline",
			"error", err,
		)
		return err
	}

	slog.Info("set read deadline successfully",
		slog.String("deadline", t.String()),
	)

	return nil
}
