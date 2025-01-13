package moqt

type ServerSession interface {
	Session
	//GoAway(string /* New Session URI */, time.Duration /* Timeout to terminate */)
}

var _ ServerSession = (*serverSession)(nil)

type serverSession struct {
	*session
}
