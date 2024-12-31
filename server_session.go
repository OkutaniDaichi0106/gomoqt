package moqt

import (
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

type ServerSession interface {
	Session
	Terminate(error)
	GoAway(string /* New Session URI */, time.Duration /* Timeout to terminate */)
}

// var _ serverSession = (*ServerSession)(nil)

type serverSession struct {
	session
}

func (sess *serverSession) GoAway(uri string, timeout time.Duration) {
	gam := message.GoAwayMessage{
		NewSessionURI: uri,
	}

	err := gam.Encode(sess.stream)
	if err != nil {
		slog.Error("failed to send GO_AWAY message", slog.String("error", err.Error()))
		return
	}

	// Wait during the given time
	time.Sleep(timeout)

	// Terminate the Session with an GO_AWAY_TIMEOUT error
	sess.Terminate(ErrGoAwayTimeout)
}
