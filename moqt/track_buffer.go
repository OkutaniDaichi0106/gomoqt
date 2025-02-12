package moqt

import (
	"context"
)

func NewTrackBuffer(config SubscribeConfig) TrackBuffer {
	return &trackBuffer{}
}

var _ TrackWriter = (*trackBuffer)(nil)
var _ TrackReader = (*trackBuffer)(nil)

type trackBuffer struct {
	path TrackPath

	queue groupBufferHeap

	latestGroupSequence GroupSequence

	ch chan struct{}

	// Config
	closed bool

	closedErr error
}

func (t *trackBuffer) TrackPath() TrackPath {
	return t.path
}

func (t *trackBuffer) GroupOrder() GroupOrder {
	return t.queue.groupOrder
}

func (t *trackBuffer) LatestGroupSequence() GroupSequence {
	return t.latestGroupSequence
}

func (t *trackBuffer) CountGroups() int {
	return t.queue.Len()
}

func (t *trackBuffer) OpenGroup(sequence GroupSequence) (GroupWriter, error) {
	gb := &GroupBuffer{sequence: sequence}
	t.queue.Push(gb)

	// Update latest group sequence
	if gb.GroupSequence() > t.latestGroupSequence {
		t.latestGroupSequence = gb.GroupSequence()
	}

	// Notify waiting routines (non-blocking)
	select {
	case t.ch <- struct{}{}:
	default:
	}

	// Notify waiting routines (non-blocking)
	select {
	case t.ch <- struct{}{}:
	default:
	}

	return nil
}

func (t *trackBuffer) AcceptGroup(ctx context.Context) (GroupReader, error) {
	for {
		if t.queue.Len() > 0 {
			return t.queue.Pop().(GroupReader), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.ch:
			// Continue loop when notification received
			continue
		}
	}
}

func (t *trackBuffer) Close() error {
	return nil
}

func (t *trackBuffer) CloseWithError(err error) error {
	return nil
}
