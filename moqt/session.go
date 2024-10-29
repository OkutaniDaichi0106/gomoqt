package moqt

type Session struct {
	Connection
	stream SessionStream
	//version Version
}

type SessionHandler interface {
	HandleSession(Session)
	SetupHandler
}
