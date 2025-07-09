package moqt

type SubscribeController interface {
	SubscribeConfig() *SubscribeConfig
	UpdateSubscribe(*SubscribeConfig) error
	Close() error
	CloseWithError(SubscribeErrorCode) error
}
