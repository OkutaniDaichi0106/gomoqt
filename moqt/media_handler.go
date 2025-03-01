package moqt

type MediaHandler interface {
	ServeTrack(w TrackWriter, r SubscribeConfig)
}
