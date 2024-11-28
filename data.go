package moqt

type DataHandler interface {
	HandleData(Group, ReceiveStream)
}

type DataHandlerFunc func(Group, ReceiveStream)

func (f DataHandlerFunc) HandleData(g Group, s ReceiveStream) {
	f(g, s)
}
