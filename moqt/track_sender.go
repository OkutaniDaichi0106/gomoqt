package moqt

import "context"

func newTrackSender(trackCtx *trackContext,
	queue *outgoingGroupStreamQueue,
	openGroupFunc func(*groupContext) (*sendGroupStream, error),
) *trackSender {
	return &trackSender{
		trackCtx:      trackCtx,
		groupQueue:    queue,
		openGroupFunc: openGroupFunc,
	}
}

var _ TrackWriter = (*trackSender)(nil)

type trackSender struct {
	trackCtx *trackContext

	groupQueue *outgoingGroupStreamQueue

	openGroupFunc func(*groupContext) (*sendGroupStream, error)
}

func (s *trackSender) OpenGroup(seq GroupSequence) (GroupWriter, error) {
	if err := s.trackCtx.Err(); err != nil {
		if reason := context.Cause(s.trackCtx); reason != nil {
			return nil, reason
		}
		return nil, err
	}

	grpCtx := newGroupContext(s.trackCtx, seq)
	stream, err := s.openGroupFunc(grpCtx)
	if err != nil {
		s.trackCtx.Logger().Error("failed to open group",
			"error", err,
		)
		return nil, err
	}

	s.groupQueue.add(stream)
	go func() {
		<-grpCtx.Done()
		s.groupQueue.remove(stream)
	}()

	return stream, nil
}

func (s *trackSender) Close() error {
	if err := s.trackCtx.Err(); err != nil {
		if reason := context.Cause(s.trackCtx); reason != nil {
			return reason
		}
		return err
	}

	s.trackCtx.cancel(ErrClosedTrack)

	s.groupQueue.clear(nil)

	return nil
}

func (s *trackSender) CloseWithError(reason error) error {
	if err := s.trackCtx.Err(); err != nil {
		if reason := context.Cause(s.trackCtx); reason != nil {
			return reason
		}
		return err
	}

	s.trackCtx.cancel(reason)

	s.groupQueue.clear(reason)

	return nil
}
