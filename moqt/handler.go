package moqt

type TrackHandler interface {
	ServeTrack(TrackWriter, SubscribeConfig)
	ServeAnnouncements(AnnouncementWriter)

	GetInfo(TrackPath) (Info, error)
}

var NotFoundHandler TrackHandler = &notFoundHandler{}

type notFoundHandler struct{}

func (h *notFoundHandler) ServeTrack(w TrackWriter, r SubscribeConfig) {
	w.CloseWithError(ErrTrackDoesNotExist)
}

func (h *notFoundHandler) GetInfo(TrackPath) (Info, error) {
	return NotFoundInfo, ErrTrackDoesNotExist
}

func (h *notFoundHandler) ServeAnnouncements(w AnnouncementWriter) {
}
