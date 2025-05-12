package moqt

// type Handler interface {
// 	ServeTrack(TrackWriter, *SubscribeConfig)
// 	ServeAnnouncements(AnnouncementWriter, *AnnounceConfig)

// 	GetInfo(TrackPath) (Info, error)
// }

type TrackHandler interface {
	ServeTrack(TrackWriter, *SubscribeConfig)
}

// type AnnouncementHandler interface {
// 	ServeAnnouncements(AnnouncementWriter, *AnnounceConfig)
// }

// type InfoHandler interface {
// 	GetInfo(TrackPath) (Info, error)
// }

var NotFoundTrackHandler TrackHandler = &notFoundHandler{}

// var NotFoundAnnouncementHandler AnnouncementHandler = &notFoundHandler{}

var _ TrackHandler = (*notFoundHandler)(nil)

// var _ AnnouncementHandler = (*notFoundHandler)(nil)

// var _ InfoHandler = (*notFoundHandler)(nil)

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
