package moqt

type Subscriber struct {
	BroadcastPath BroadcastPath

	TrackName TrackName

	TrackReader TrackReader

	SubscribeStream *sendSubscribeStream
}
