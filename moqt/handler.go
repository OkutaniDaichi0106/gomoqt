package moqt

type Handler interface {
	ServeTrack(TrackWriter, *SubscribeConfig)
	ServeAnnouncements(AnnouncementWriter, *AnnounceConfig)

	GetInfo(TrackPath) (Info, error)
}

type TrackHandler interface {
	ServeTrack(TrackWriter, *SubscribeConfig)
}

type AnnouncementHandler interface {
	ServeAnnouncements(AnnouncementWriter, *AnnounceConfig)
}

type InfoHandler interface {
	GetInfo(TrackPath) (Info, error)
}

var NotFoundHandler Handler = &notFoundHandler{}

type notFoundHandler struct{}

func (h *notFoundHandler) ServeTrack(w TrackWriter, config *SubscribeConfig) {
	w.CloseWithError(ErrTrackDoesNotExist)
}

func (h *notFoundHandler) GetInfo(TrackPath) (Info, error) {
	return NotFoundInfo, ErrTrackDoesNotExist
}

func (h *notFoundHandler) ServeAnnouncements(w AnnouncementWriter, config *AnnounceConfig) {
	// Do nothing
}
