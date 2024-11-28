package moqt

import (
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

type ServerSession struct {
	*session
}

func (sess *ServerSession) GoAway(uri string, timeout time.Duration) {
	gam := message.GoAwayMessage{
		NewSessionURI: uri,
	}

	_, err := sess.sessStr.Write(gam.SerializePayload())
	if err != nil {
		slog.Error("failed to send GO_AWAY message", slog.String("error", err.Error()))
		return
	}

	// Wait during the given time
	time.Sleep(timeout)

	//
	if len(sess.receivedSubscriptions) != 0 {
		slog.Info("subscription is still on the Session")

		// Terminate the Session with an GO_AWAY_TIMEOUT error
		sess.Terminate(ErrGoAwayTimeout)

		return
	}
}
