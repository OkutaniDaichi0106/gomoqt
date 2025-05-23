package moqt

type SentSubscription interface {
	SubscribeID() SubscribeID
	SubuscribeConfig() *SubscribeConfig
	// TrackPath() TrackPath
	UpdateSubscribe(new *SubscribeConfig) error
}
