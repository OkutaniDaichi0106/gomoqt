package moqt

type Subscription struct {
	BroadcastPath BroadcastPath
	TrackName     TrackName

	TrackReader TrackReader

	Controller SubscribeController
}
