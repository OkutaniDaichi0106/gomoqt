package moqt

type TrackHandler interface {
	ServeTrack(*Publication)
}

var NotFound = func(pub *Publication) {
	if pub == nil {
		return
	}
	if pub.TrackWriter == nil {
		return
	}
	pub.Controller.CloseWithError(TrackNotFoundErrorCode)
}

var NotFoundHandler TrackHandler = TrackHandlerFunc(NotFound)

type TrackHandlerFunc func(*Publication)

func (f TrackHandlerFunc) ServeTrack(pub *Publication) {
	f(pub)
}
