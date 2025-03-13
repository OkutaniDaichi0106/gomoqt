package internal

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

type SessionStream struct {
	Stream transport.Stream

	SessionClientMessage message.SessionClientMessage
	SessionServerMessage message.SessionServerMessage
}

func (ss *SessionStream) SendSessionUpdateMessage(sum message.SessionUpdateMessage) error {
	_, err := sum.Encode(ss.Stream)
	if err != nil {
		slog.Error("failed to send a SESSION_UPDATE message", "error", err)
		return err
	}

	slog.Debug("sent a SESSION_UPDATE message")

	return nil
}
