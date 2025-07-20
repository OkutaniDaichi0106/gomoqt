package moqt

type TrackHandler interface {
	ServeTrack(*TrackWriter)
}

var NotFound = func(tw *TrackWriter) {
	if tw == nil {
		return
	}

	tw.CloseWithError(TrackNotFoundErrorCode)
}

var NotFoundHandler TrackHandler = TrackHandlerFunc(NotFound)

type TrackHandlerFunc func(*TrackWriter)

func (f TrackHandlerFunc) ServeTrack(tw *TrackWriter) {
	f(tw)
}
