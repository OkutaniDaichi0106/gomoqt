package moqt

import (
	"container/heap"
	"context"
	"fmt"
	"io"
)

var _ TrackReader = (*trackBufferReader)(nil)

func newTrackBufferReader(tb *TrackBuffer, config *SubscribeConfig) *trackBufferReader {
	if config == nil {
		config = &SubscribeConfig{}
	}

	tr := &trackBufferReader{
		trackBuffer:  tb,
		config:       config,
		sequenceCh:   tb.addNotifyChannel(),
		sequenceHeap: newGroupSequenceHeap(config.GroupOrder, config.MinGroupSequence, config.MaxGroupSequence),
	}

	return tr
}

type trackBufferReader struct {
	trackBuffer *TrackBuffer

	config *SubscribeConfig

	sequenceCh   chan GroupSequence
	sequenceHeap *groupSequenceHeap

	canceled    bool
	canceledErr error
}

func (tr *trackBufferReader) TrackPath() TrackPath {
	return tr.trackBuffer.TrackPath()
}

func (tr *trackBufferReader) LatestGroupSequence() GroupSequence {
	return tr.trackBuffer.LatestGroupSequence()
}

func (tr *trackBufferReader) Info() Info {
	return tr.trackBuffer.Info()
}

func (tr *trackBufferReader) AcceptGroup(ctx context.Context) (GroupReader, error) {
	// Return EOF if the track is closed and there are no more groups.
	if len(tr.trackBuffer.groupMap) <= 0 && tr.trackBuffer.closed {
		return nil, io.EOF
	}

	for {
		// Check for reader cancellation
		if tr.canceled {
			if tr.canceledErr != nil {
				return nil, fmt.Errorf("track already unsubscribed with error: %w", tr.canceledErr)
			}
			return nil, ErrUnsubscribedTrack
		}

		// If the heap is not empty, pop the GroupSequence from the heap.
		if tr.sequenceHeap.Len() > 0 {
			// Pop the GroupSequence from the heap.
			seq := heap.Pop(tr.sequenceHeap).(GroupSequence)

			// Get a group with the group sequence.
			gb, ok := tr.trackBuffer.getGroup(seq)
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
			heap.Push(tr.sequenceHeap, seq)
		}
	}
}

func (tr *trackBufferReader) Close() error {
	if tr.canceled {
		if tr.canceledErr == nil {
			return fmt.Errorf("track already closed with error: %w", tr.canceledErr)
		}
		return ErrClosedTrack
	}

	// Close the sequence channel and remove it from the track buffer.
	close(tr.sequenceCh)
	tr.trackBuffer.removeNotifyChannel(tr.sequenceCh)
	tr.sequenceCh = nil

	tr.canceled = true
	tr.sequenceHeap = nil
	tr.trackBuffer = nil
	tr.canceledErr = nil

	return nil
}

func (tr *trackBufferReader) CloseWithError(err error) error {
	if tr.canceled {
		if tr.canceledErr == nil {
			return fmt.Errorf("track already closed with error: %w", tr.canceledErr)
		}
		return ErrClosedTrack
	}

	close(tr.sequenceCh)
	tr.trackBuffer.removeNotifyChannel(tr.sequenceCh)
	tr.sequenceCh = nil

	tr.canceled = true
	tr.sequenceHeap = nil
	tr.trackBuffer = nil
	tr.canceledErr = err

	return nil
}
