package moqt

type ReceivedSubscription interface {
	SubscribeID() SubscribeID
	// TrackPath() TrackPath
	SubuscribeConfig() *SubscribeConfig
	Updated() <-chan struct{}
}
