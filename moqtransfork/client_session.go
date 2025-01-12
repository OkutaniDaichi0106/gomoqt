package moqtransfork

type ClientSession interface {
	Session
}

var _ ClientSession = (*clientSession)(nil)

type clientSession struct {
	/*
	 * session
	 */
	*session
}
