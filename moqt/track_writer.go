package moqt

import "context"

type TrackWriter interface {
	// Create a new group writer
	OpenGroup(GroupSequence) (GroupWriter, error)

	Close() error

	CloseWithError(error) error
}

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
	if s.trackCtx.Err() != nil {
		if reason := context.Cause(s.trackCtx); reason != nil {
			return nil, reason
		}
		return nil, ErrClosedGroup
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
	if s.trackCtx.Err() != nil {
		if reason := context.Cause(s.trackCtx); reason != nil {
			return reason
		}
		return ErrClosedGroup
	}

	s.trackCtx.cancel(ErrClosedGroup)

	return nil
}

func (s *trackSender) CloseWithError(reason error) error {
	if s.trackCtx.Err() != nil {
		if reason := context.Cause(s.trackCtx); reason != nil {
			return reason
		}
		return ErrClosedGroup
	}

	s.trackCtx.cancel(reason)

	return nil
}
