package moqt

type TrackHandler interface {
	ServeTrack(TrackWriter, SendTrackStream)
}

var NotFound = func(w TrackWriter, sub SendTrackStream) {
	w.CloseWithError(ErrTrackDoesNotExist)
}
var NotFoundHandler TrackHandler = TrackHandlerFunc(NotFound)

type TrackHandlerFunc func(TrackWriter, SendTrackStream)

func (f TrackHandlerFunc) ServeTrack(w TrackWriter, sub SendTrackStream) {
	f(w, sub)
}
