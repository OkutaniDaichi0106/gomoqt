package moqt

type clientSession interface {
	Publisher() *Publisher
	Subscriber() *Subscriber
	Terminate(error)
}

var _ clientSession = (*ClientSession)(nil)

type ClientSession struct {
	/*
	 * session
	 */
	session
}
