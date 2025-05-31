package moqt

import (
	"context"
	"errors"
)

func newTrackReceiver(ctx *trackContext, queue *incomingGroupStreamQueue) *trackReceiver {
	return &trackReceiver{
		trackCtx:   ctx,
		groupQueue: queue,
	}
}

var _ TrackReader = (*trackReceiver)(nil)

type trackReceiver struct {
	trackCtx *trackContext

	groupQueue *incomingGroupStreamQueue
}

func (r *trackReceiver) AcceptGroup(ctx context.Context) (GroupReader, error) {
	return r.groupQueue.dequeue(ctx)
}

func (r *trackReceiver) Close() error {
	r.trackCtx.cancel(ErrClosedTrack)
	return nil
}

func (r *trackReceiver) CloseWithError(reason error) error {
	if reason == nil {
		reason = ErrInternalError
	}

	r.trackCtx.cancel(reason)

	var grperr GroupError
	if !errors.As(reason, &grperr) {
		grperr = ErrInternalError.WithReason(reason.Error())
	}
	r.groupQueue.clear(grperr)

	return nil

}
