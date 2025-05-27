package moqt

import "context"

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
	defer r.trackCtx.cancel(ErrClosedTrack)
	return nil
}

func (r *trackReceiver) CloseWithError(err error) error {
	defer r.trackCtx.cancel(ErrClosedTrack)

	return nil

}
