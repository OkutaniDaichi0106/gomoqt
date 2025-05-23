package moqt

type TrackHandler interface {
	ServeTrack(TrackWriter, ReceivedSubscription)
}

var NotFound = func(w TrackWriter, sub ReceivedSubscription) {
	w.CloseWithError(ErrTrackDoesNotExist)
}
var NotFoundHandler TrackHandler = TrackHandlerFunc(NotFound)

type TrackHandlerFunc func(TrackWriter, ReceivedSubscription)

func (f TrackHandlerFunc) ServeTrack(w TrackWriter, sub ReceivedSubscription) {
	f(w, sub)
}
