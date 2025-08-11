package moqt

import (
	"context"
	"sync"
)

var DefaultRouter *Router = defaultRouter

var defaultRouter *Router = NewRouter()

func HandleFunc(pattern string, f func(w ResponseWriter, r *Request)) {
	DefaultRouter.HandleFunc(pattern, f)
}

func Handle(pattern string, h Handler) {
	DefaultRouter.Handle(pattern, h)
}

func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]Handler),
	}
}

type Router struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

func (r *Router) Handle(pattern string, h Handler) {
	r.register(pattern, h)
}

func (r *Router) HandleFunc(pattern string, f func(w ResponseWriter, r *Request)) {
	r.Handle(pattern, HandlerFunc(f))
}

func (r *Router) register(path string, h Handler) {
	if path == "" {
		panic("moq: path cannot be empty")
	}
	if path[0] != '/' {
		panic("moq: path must start with '/'")
	}
	if h == nil {
		panic("moq: handler cannot be nil")
	}
	if f, ok := h.(HandlerFunc); ok && f == nil {
		panic("moq: handler function cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handlers == nil {
		r.handlers = make(map[string]Handler)
	}
	r.handlers[path] = h
}

func (r *Router) Handler(pattern string) Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if h, ok := r.handlers[pattern]; ok {
		return h
	}
	return NotFoundHandler
}

func (r *Router) ServeMOQ(w ResponseWriter, req *Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler := r.Handler(req.Path)
	if handler == nil {
		handler = NotFoundHandler
	}

	handler.ServeMOQ(w, req)
}

type Handler interface {
	ServeMOQ(w ResponseWriter, r *Request)
}

type Request struct {
	Path       string
	Versions   []Version
	Extensions *Parameters

	ctx context.Context
}

func (r *Request) Context() context.Context {
	return r.ctx
}

type ResponseWriter interface {
	Accept(v Version, extensions *Parameters) error
	Reject(code SessionErrorCode) error
}

var _ Handler = (*HandlerFunc)(nil)

type HandlerFunc func(w ResponseWriter, r *Request)

func (f HandlerFunc) ServeMOQ(w ResponseWriter, r *Request) {
	f(w, r)
}

var NotFoundFunc func(w ResponseWriter, r *Request) = func(w ResponseWriter, r *Request) {
	w.Reject(SetupFailedErrorCode)
}

var NotFoundHandler Handler = HandlerFunc(NotFoundFunc)
