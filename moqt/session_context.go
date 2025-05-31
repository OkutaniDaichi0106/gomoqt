package moqt

import (
	"context"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/moqtrace"
)

func newSessionContext(connCtx context.Context,
	version protocol.Version,
	path string,
	clientParams *Parameters,
	serverParams *Parameters,
	logger *slog.Logger,
	tracer *moqtrace.SessionTracer,
) *sessionContext {
	ctx, cancel := context.WithCancelCause(connCtx)

	var sessionLogger *slog.Logger
	if logger != nil {
		sessionLogger = logger.With(slog.String("remote_address", "TODO"))
	}

	return &sessionContext{
		Context:          ctx,
		cancel:           cancel,
		path:             path,
		clientParameters: clientParams,
		serverParameters: serverParams,
		version:          version,
		logger:           sessionLogger,
		tracer:           tracer,
	}
}

var _ context.Context = (*sessionContext)(nil)

type sessionContext struct {
	context.Context
	cancel context.CancelCauseFunc

	wg sync.WaitGroup // WaitGroup for session cleanup

	path string

	// Version of the protocol used in this session
	version protocol.Version

	// Parameters specified by the client and server
	clientParameters *Parameters

	// Parameters specified by the server
	serverParameters *Parameters

	// bitrate atomic.Uint64 // Bitrate in bits per second

	logger *slog.Logger

	tracer *moqtrace.SessionTracer
}

func (sc *sessionContext) Logger() *slog.Logger {
	return sc.logger
}

func (sc *sessionContext) Path() string {
	return sc.path
}

func (sc *sessionContext) Version() protocol.Version {
	return sc.version
}

func (sc *sessionContext) ClientParameters() *Parameters {
	if sc.clientParameters == nil {
		return NewParameters()
	}
	return sc.clientParameters
}

func (sc *sessionContext) ServerParameters() *Parameters {
	if sc.serverParameters == nil {
		return NewParameters()
	}
	return sc.serverParameters
}

func (sc *sessionContext) Tracer() *moqtrace.SessionTracer {
	return sc.tracer
}
