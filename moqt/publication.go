package moqt

type Publication struct {
	BroadcastPath BroadcastPath
	TrackName     TrackName
	SubscribeID   SubscribeID

	TrackWriter TrackWriter

	Controller PublishController
}
