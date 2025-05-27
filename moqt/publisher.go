package moqt

type Publisher struct {
	BroadcastPath BroadcastPath

	TrackName TrackName

	TrackWriter TrackWriter

	SubscribeStream ReceiveSubscribeStream
}
