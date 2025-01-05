package moqt

type ServerSession interface {
	Session
	//GoAway(string /* New Session URI */, time.Duration /* Timeout to terminate */)
}

var _ ServerSession = (*serverSession)(nil)

type serverSession struct {
	session
}

// func (sess *serverSession) GoAway(uri string, timeout time.Duration) {
// 	gam := message.GoAwayMessage{
// 		NewSessionURI: uri,
// 	}

// 	err := gam.Encode(sess.stream)
// 	if err != nil {
// 		slog.Error("failed to send GO_AWAY message", slog.String("error", err.Error()))
// 		return
// 	}

// 	// Wait during the given time
// 	time.Sleep(timeout)

// 	// TODO:

// 	// Terminate the Session with an GO_AWAY_TIMEOUT error
// 	sess.Terminate(ErrGoAwayTimeout)
// }
