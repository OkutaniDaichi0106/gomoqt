package moqt

type ReceivedSubscription interface {
	SubscribeID() SubscribeID
	TrackName() string
	SubuscribeConfig() *SubscribeConfig
	Updated() <-chan struct{}
}
