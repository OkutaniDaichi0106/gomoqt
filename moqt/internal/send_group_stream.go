package internal

import (
	"errors"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newSendGroupStream(gm *message.GroupMessage, stream transport.SendStream) *SendGroupStream {
	return &SendGroupStream{
		GroupMessage: *gm,
		SendStream:   stream,
		startTime:    time.Now(),
	}
}

type SendGroupStream struct {
	GroupMessage message.GroupMessage
	SendStream   transport.SendStream
	startTime    time.Time // TODO: Delete if not used
}

func (sgs *SendGroupStream) SendFrameBytes(frame []byte) error {
	fm := message.FrameMessage{
		Payload: frame,
	}
	_, err := fm.Encode(sgs.SendStream)
	if err != nil {
		slog.Error("failed to write a frame message",
			"stream_id", sgs.SendStream.StreamID(),
			"error", err,
		)

		// Signal the group error code
		var grperr GroupError
		if !errors.As(err, &grperr) {
			grperr = ErrInternalError
		}

		code := transport.StreamErrorCode(grperr.GroupErrorCode())

		sgs.SendStream.CancelWrite(code)

		return err
	}

	slog.Debug("wrote a frame message",
		slog.Any("stream_id", sgs.SendStream.StreamID()),
	)

	return nil
}

func (sgs *SendGroupStream) StartAt() time.Time {
	return sgs.startTime
}

func (sgs *SendGroupStream) Close() error {
	err := sgs.SendStream.Close()
	if err != nil {
		slog.Error("failed to close a send stream",
			"error", err,
			"stream_id", sgs.SendStream.StreamID(),
		)
		return err
	}

	slog.Debug("closed a send group stream gracefully",
		"stream_id", sgs.SendStream.StreamID(),
	)

	return nil
}

func (sgs *SendGroupStream) SetWriteDeadline(t time.Time) error {
	err := sgs.SendStream.SetWriteDeadline(t)
	if err != nil {
		slog.Error("failed to set write deadline",
			"stream_id", sgs.SendStream.StreamID(),
			"error", err,
		)
		return err
	}

	return nil
}

func (sgs *SendGroupStream) CancelWrite(code protocol.GroupErrorCode) {
	sgs.SendStream.CancelWrite(transport.StreamErrorCode(code))

	slog.Debug("canceled write",
		"code", code,
		"stream_id", sgs.SendStream.StreamID(),
	)
}
