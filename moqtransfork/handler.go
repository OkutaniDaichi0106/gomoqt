package moqtransfork

import (
	"strings"
	"sync"
)

type HandlerFunc func(ServerSession)

var NotFoundFunc HandlerFunc = func(ServerSession) {}

var DefaultHandler *ServeMux = NewServeMux()

func NewServeMux() *ServeMux {
	return &ServeMux{
		handlerFuncs: make(map[string]HandlerFunc),
	}
}

type ServeMux struct {
	mu sync.Mutex

	/*
	 * Path pattern -> HandlerFunc
	 */
	handlerFuncs map[string]HandlerFunc
}

func (h *ServeMux) HandlerFunc(pattern string, op HandlerFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !strings.HasPrefix(pattern, "/") {
		panic("invalid path: path should start with \"/\"")
	}

	h.handlerFuncs[pattern] = op
}

func (mux *ServeMux) findHandlerFunc(pattern string) HandlerFunc {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	handlerFunc, ok := mux.handlerFuncs[pattern]

	if !ok {
		return NotFoundFunc
	}

	return handlerFunc
}

func HandleFunc(pattern string, op func(ServerSession)) {
	DefaultHandler.HandlerFunc(pattern, op)
}
