package moqt

type Subscription struct {
	BroadcastPath BroadcastPath
	TrackName     TrackName
	SubscribeID   SubscribeID

	TrackReader TrackReader

	Controller SubscribeController
}
