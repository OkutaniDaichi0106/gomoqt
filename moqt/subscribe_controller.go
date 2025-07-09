package moqt

type SubscribeController interface {
	SubscribeID() SubscribeID // TODO: Should I include this in context?
	SubscribeConfig() *SubscribeConfig
	UpdateSubscribe(*SubscribeConfig) error
	Close() error
	CloseWithError(SubscribeErrorCode) error
}
