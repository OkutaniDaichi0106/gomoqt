package moqt

type TrackResolver interface {
	ServeTrack(TrackWriter, SubscribeConfig)
	ServeAnnouncements(AnnouncementWriter)

	GetInfo(TrackPath) (Info, error)
}

var NotFoundHandler TrackResolver = &notFoundHandler{}

type notFoundHandler struct{}

func (h *notFoundHandler) ServeTrack(w TrackWriter, r SubscribeConfig) {
	w.CloseWithError(ErrTrackDoesNotExist)
}

func (h *notFoundHandler) GetInfo(TrackPath) (Info, error) {
	return Info{}, ErrTrackDoesNotExist
}

func (h *notFoundHandler) ServeAnnouncements(w AnnouncementWriter) {
}

// var DefaultHandler *ServeMux = NewServeMux()

// func NewServeMux() *ServeMux {
// 	return &ServeMux{
// 		handlerFuncs: make(map[string]HandlerFunc),
// 	}
// }

// type ServeMux struct {
// 	mu sync.Mutex

// 	/*
// 	 * Path pattern -> HandlerFunc
// 	 */
// 	handlerFuncs map[string]HandlerFunc
// }

// func (h *ServeMux) HandlerFunc(pattern string, op func(Session)) {
// 	h.mu.Lock()
// 	defer h.mu.Unlock()

// 	if !strings.HasPrefix(pattern, "/") {
// 		panic("invalid path: path should start with \"/\"")
// 	}

// 	h.handlerFuncs[pattern] = op
// }

// func (mux *ServeMux) findHandlerFunc(pattern string) HandlerFunc {
// 	mux.mu.Lock()
// 	defer mux.mu.Unlock()

// 	handlerFunc, ok := mux.handlerFuncs[pattern]

// 	if !ok {
// 		return NotFoundFunc
// 	}

// 	return handlerFunc
// }

// func HandleFunc(pattern string, op func(Session)) {
// 	DefaultHandler.HandlerFunc(pattern, op)
// }
