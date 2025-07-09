package moqt

type PublishController interface {
	WriteInfo(Info) error
	SubscribeID() SubscribeID // TODO: Should I include this in context?
	SubscribeConfig() (*SubscribeConfig, error)
	Updated() <-chan struct{}
	Close() error
	CloseWithError(SubscribeErrorCode) error
}
