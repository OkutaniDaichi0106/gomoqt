package moqt

type Publication struct {
	BroadcastPath BroadcastPath
	TrackName     TrackName

	TrackWriter TrackWriter

	Controller PublishController
}
