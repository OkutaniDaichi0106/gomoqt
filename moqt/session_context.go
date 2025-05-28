package moqt

import (
	"context"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/moqtrace"
)

func newSessionContext(parentCtx context.Context, path string, logger *slog.Logger) *sessionContext {
	ctx, cancel := context.WithCancelCause(parentCtx)
	return &sessionContext{
		Context: ctx,
		cancel:  cancel,
		logger:  logger.With(slog.String("remote_address", "session")),
		path:    path,
	}
}

var _ context.Context = (*sessionContext)(nil)

type sessionContext struct {
	context.Context
	cancel context.CancelCauseFunc

	path string

	// trackCtxs map[*trackContext]struct{}

	version protocol.Version

	// Parameters specified by the client and server
	clientParameters *Parameters

	// Parameters specified by the server
	serverParameters *Parameters

	logger *slog.Logger

	tracer moqtrace.SessionTracer
}

func (sc *sessionContext) setup(version protocol.Version, clientParams *Parameters, serverParams *Parameters) {
	sc.version = version
	sc.clientParameters = clientParams
	sc.serverParameters = serverParams
}

func (sc *sessionContext) Logger() *slog.Logger {
	return sc.logger
}
