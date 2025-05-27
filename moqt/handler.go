package moqt

type TrackHandler interface {
	ServeTrack(*Publisher)
}

var NotFound = func(pub *Publisher) {
	if pub == nil {
		return
	}
	if pub.TrackWriter == nil {
		return
	}
	pub.TrackWriter.CloseWithError(ErrTrackDoesNotExist)
}

var NotFoundHandler TrackHandler = TrackHandlerFunc(NotFound)

type TrackHandlerFunc func(*Publisher)

func (f TrackHandlerFunc) ServeTrack(pub *Publisher) {
	f(pub)
}
