package moqt

import (
	"container/heap"
	"context"
)

var _ TrackReader = (*trackBufferReader)(nil)

func newTrackBufferReader(tb *TrackBuffer, config *SubscribeConfig) *trackBufferReader {
	if config == nil {
		config = &SubscribeConfig{}
	}

	tr := &trackBufferReader{
		buffer:     tb,
		config:     config,
		sequenceCh: tb.addNotifyChannel(),
		heap:       newGroupSequenceHeap(config.GroupOrder),
	}

	return tr
}

type trackBufferReader struct {
	buffer *TrackBuffer

	config *SubscribeConfig

	sequenceCh chan GroupSequence
	heap       *groupSequenceHeap

	closed    bool
	closedErr error
}

func (tr *trackBufferReader) AcceptGroup(ctx context.Context) (GroupReader, error) {
	// Return EOF if the track is closed and there are no more groups.
	if len(tr.buffer.groupMap) <= 0 && tr.buffer.closed.Load() {
		if tr.buffer.closedErr != nil {
			return nil, tr.buffer.closedErr
		}
		return nil, ErrClosedGroup
	}

	for {
		// Check for reader cancellation
		if tr.closed {
			if tr.closedErr != nil {
				return nil, tr.closedErr
			}
			return nil, ErrUnsubscribedTrack // TODO:
		}

		// If the heap is not empty, pop the GroupSequence from the heap.
		if tr.heap.Len() > 0 {
			// Pop the GroupSequence from the heap.
			seq := heap.Pop(tr.heap).(GroupSequence)

			// Get a group with the group sequence.
			gb, ok := tr.buffer.getGroup(seq)
			if !ok {
				continue
			}

			return newGroupBufferReader(gb), nil
		}

		// Wait for a new GroupSequence to be added or the context to be canceled.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case seq, ok := <-tr.sequenceCh:
			// If the channel is closed, return an error.
			if !ok {
				return nil, ErrClosedTrack
			}

			// Enqueue the GroupSequence to the heap.
			heap.Push(tr.heap, seq)
		}
	}
}

func (tr *trackBufferReader) Close() error {
	if tr.closed {
		if tr.closedErr != nil {
			return tr.closedErr
		}
		return ErrClosedTrack
	}

	// Close the sequence channel and remove it from the track buffer.
	close(tr.sequenceCh)
	tr.buffer.removeNotifyChannel(tr.sequenceCh)
	tr.sequenceCh = nil

	tr.closed = true
	tr.heap = nil
	tr.buffer = nil
	tr.closedErr = nil

	return nil
}

func (tr *trackBufferReader) CloseWithError(err error) error {
	if tr.closed {
		if tr.closedErr != nil {
			return tr.closedErr
		}
		return ErrClosedTrack
	}

	close(tr.sequenceCh)
	tr.buffer.removeNotifyChannel(tr.sequenceCh)
	tr.sequenceCh = nil

	tr.closed = true
	tr.heap = nil
	tr.buffer = nil
	tr.closedErr = err

	return nil
}
