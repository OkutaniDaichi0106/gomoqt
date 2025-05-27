package moqt

import (
	"context"
	"log/slog"
)

func newTrackContext(sessCtx *sessionContext, id SubscribeID, path BroadcastPath, name TrackName) *trackContext {
	var logger *slog.Logger
	if sessLogger := sessCtx.Logger(); sessLogger != nil {
		logger = sessLogger.With("subscribe_id", id.String(),
			"broadcast_path", path.String(),
			"track_name", name,
		)
	}

	ctx, cancel := context.WithCancelCause(sessCtx)

	return &trackContext{
		Context: ctx,
		cancel:  cancel,

		id:     id,
		path:   path,
		name:   name,
		logger: logger,
	}
}

type trackContext struct {
	context.Context
	cancel context.CancelCauseFunc

	id   SubscribeID
	path BroadcastPath
	name TrackName

	logger *slog.Logger
}

func (tc *trackContext) Logger() *slog.Logger {
	return tc.logger
}
