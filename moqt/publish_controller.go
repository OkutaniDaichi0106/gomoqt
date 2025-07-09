package moqt

type PublishController interface {
	WriteInfo(Info) error
	SubscribeConfig() (*SubscribeConfig, error)
	Updated() <-chan struct{}
	Close() error
	CloseWithError(SubscribeErrorCode) error
}
