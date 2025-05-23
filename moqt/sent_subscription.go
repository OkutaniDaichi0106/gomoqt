package moqt

type SentSubscription interface {
	SubscribeID() SubscribeID
	TrackName() string
	SubuscribeConfig() *SubscribeConfig
	UpdateSubscribe(new *SubscribeConfig) error
}
