package moqt

import (
	"context"
	"log/slog"
)

func newGroupContext(trackCtx *trackContext, seq GroupSequence) *groupContext {
	ctx, cancel := context.WithCancelCause(trackCtx)

	var logger *slog.Logger
	if trackCtx.logger != nil {
		logger = trackCtx.logger.With(
			"group_sequence", seq.String(),
		)
	}

	return &groupContext{
		Context: ctx,
		cancel:  cancel,
		logger:  logger,
		seq:     seq,
	}
}

type groupContext struct {
	context.Context
	cancel context.CancelCauseFunc

	seq GroupSequence

	logger *slog.Logger
}

func (g *groupContext) Logger() *slog.Logger {
	return g.logger
}

type Context interface {
	context.Context
	Logger() *slog.Logger
}
