package moqt

type TrackHandler interface {
	HandleTrack(TrackWriter, ReceivedSubscription)
	Info() (Info, bool)
}

var NotFoundTrackHandler TrackHandler = nil

// func NewTrackHandler(info Info, handler func(TrackWriter, *SubscribeConfig)) TrackHandler {
// 	return &trackHandler{
// 		info:    info,
// 		handler: handler,
// 	}
// }

var _ TrackHandler = (*defaultNotFoundTrackHandler)(nil)

type defaultNotFoundTrackHandler struct {
}

func (h *defaultNotFoundTrackHandler) HandleTrack(w TrackWriter, sub ReceivedSubscription) {
	w.CloseWithError(ErrTrackDoesNotExist)
}

func (h *defaultNotFoundTrackHandler) Info() (Info, bool) {
	return Info{}, false
}
