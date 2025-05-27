package moqt

import (
	"context"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

func newSessionContext(parentCtx context.Context, logger *slog.Logger) *sessionContext {
	ctx, cancel := context.WithCancelCause(parentCtx)
	return &sessionContext{
		Context: ctx,
		cancel:  cancel,
		logger:  logger.With(slog.String("remote_address", "session")),
	}
}

var _ context.Context = (*sessionContext)(nil)

type sessionContext struct {
	context.Context
	cancel context.CancelCauseFunc

	// trackCtxs map[*trackContext]struct{}

	version protocol.Version

	// Parameters specified by the client and server
	clientParameters *Parameters

	// Parameters specified by the server
	serverParameters *Parameters

	logger *slog.Logger
}

func (sc *sessionContext) Logger() *slog.Logger {
	return sc.logger
}
