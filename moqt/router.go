package moqt

import (
	"context"
	"sync"
)

var DefaultRouter *SetupRouter = defaultRouter

var defaultRouter *SetupRouter = NewRouter()

func HandleFunc(pattern string, f func(w SetupResponseWriter, r *SetupRequest)) {
	DefaultRouter.HandleFunc(pattern, f)
}

func Handle(pattern string, h SetupHandler) {
	DefaultRouter.Handle(pattern, h)
}

func NewRouter() *SetupRouter {
	return &SetupRouter{
		handlers: make(map[string]SetupHandler),
	}
}

type SetupRouter struct {
	mu       sync.RWMutex
	handlers map[string]SetupHandler
}

func (r *SetupRouter) Handle(pattern string, h SetupHandler) {
	r.register(pattern, h)
}

func (r *SetupRouter) HandleFunc(pattern string, f func(w SetupResponseWriter, r *SetupRequest)) {
	r.Handle(pattern, SetupHandlerFunc(f))
}

func (r *SetupRouter) register(path string, h SetupHandler) {
	if path == "" {
		panic("moq: path cannot be empty")
	}
	if path[0] != '/' {
		panic("moq: path must start with '/'")
	}
	if h == nil {
		panic("moq: handler cannot be nil")
	}
	if f, ok := h.(SetupHandlerFunc); ok && f == nil {
		panic("moq: handler function cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.handlers == nil {
		r.handlers = make(map[string]SetupHandler)
	}
	r.handlers[path] = h
}

func (r *SetupRouter) Handler(pattern string) SetupHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if h, ok := r.handlers[pattern]; ok {
		return h
	}
	return RejectSetupHandler
}

func (r *SetupRouter) ServeMOQ(w SetupResponseWriter, req *SetupRequest) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler := r.Handler(req.Path)
	if handler == nil {
		handler = RejectSetupHandler
	}

	handler.ServeMOQ(w, req)
}

type SetupHandler interface {
	ServeMOQ(w SetupResponseWriter, r *SetupRequest)
}

type SetupRequest struct {
	Path             string
	Versions         []Version
	ClientExtensions *Parameters

	ctx context.Context
}

func (r *SetupRequest) Context() context.Context {
	return r.ctx
}

type SetupResponseWriter interface {
	SelectVersion(v Version) error
	SetExtensions(extensions *Parameters)
	Reject(code SessionErrorCode) error
}

var _ SetupHandler = (*SetupHandlerFunc)(nil)

type SetupHandlerFunc func(w SetupResponseWriter, r *SetupRequest)

func (f SetupHandlerFunc) ServeMOQ(w SetupResponseWriter, r *SetupRequest) {
	f(w, r)
}

var RejectSetupFunc func(w SetupResponseWriter, r *SetupRequest) = func(w SetupResponseWriter, r *SetupRequest) {
	w.Reject(SetupFailedErrorCode)
}

var RejectSetupHandler SetupHandler = SetupHandlerFunc(RejectSetupFunc)
