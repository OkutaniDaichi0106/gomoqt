package moqt

import (
	"container/heap"
	"context"
	"fmt"
	"io"
)

var _ TrackReader = (*trackBufferReader)(nil)

func newTrackBufferReader(tb *TrackBuffer, order GroupOrder) *trackBufferReader {
	tr := &trackBufferReader{
		trackBuffer:  tb,
		sequenceCh:   tb.addNotifyChannel(),
		sequenceHeap: newGroupSequenceHeap(order),
		ch:           make(chan struct{}),
	}

	go tr.listenSequences()

	return tr
}

type trackBufferReader struct {
	trackBuffer *TrackBuffer
	priority    TrackPriority
	order       GroupOrder

	sequenceCh   chan GroupSequence
	sequenceHeap *groupSequenceHeap
	ch           chan struct{}

	closed   bool
	closeErr error
}

func (tr *trackBufferReader) TrackPath() TrackPath {
	return tr.trackBuffer.TrackPath()
}

func (tr *trackBufferReader) TrackPriority() TrackPriority {
	return tr.priority
}

func (tr *trackBufferReader) GroupOrder() GroupOrder {
	return tr.order
}

func (tr *trackBufferReader) LatestGroupSequence() GroupSequence {
	return tr.trackBuffer.LatestGroupSequence()
}

func (tr *trackBufferReader) Info() Info {
	return Info{
		TrackPriority:       tr.priority,
		LatestGroupSequence: tr.LatestGroupSequence(),
		GroupOrder:          tr.order,
	}
}

func (tr *trackBufferReader) AcceptGroup(ctx context.Context) (GroupReader, error) {
	if len(tr.trackBuffer.groupMap) <= 0 && tr.trackBuffer.closed {
		return nil, io.EOF
	}

	for {
		if tr.closed {
			if tr.closeErr != nil {
				return nil, tr.closeErr
			}
			return nil, ErrClosedTrack
		}

		if tr.sequenceHeap.Len() > 0 {
			seq := heap.Pop(tr.sequenceHeap).(GroupSequence)

			gb, ok := tr.trackBuffer.getGroup(seq)
			if !ok {
				continue
			}

			return newGroupBufferReader(gb), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-tr.ch:
		}
	}
}

func (tr *trackBufferReader) Close() error {
	if tr.closed {
		if tr.closeErr == nil {
			return fmt.Errorf("track already closed with error: %w", tr.closeErr)
		}
		return ErrClosedTrack
	}

	close(tr.sequenceCh)
	tr.trackBuffer.removeNotifyChannel(tr.sequenceCh)

	tr.closed = true
	tr.sequenceCh = nil
	tr.sequenceHeap = nil
	tr.ch = nil
	tr.trackBuffer = nil
	tr.priority = 0
	tr.order = 0
	tr.closeErr = nil

	return nil
}

func (tr *trackBufferReader) CloseWithError(err error) error {
	if tr.closed {
		if tr.closeErr == nil {
			return fmt.Errorf("track already closed with error: %w", tr.closeErr)
		}
		return ErrClosedTrack
	}

	close(tr.sequenceCh)
	tr.closeErr = err
	tr.trackBuffer.removeNotifyChannel(tr.sequenceCh)

	tr.closed = true
	tr.sequenceCh = nil
	tr.sequenceHeap = nil
	tr.ch = nil
	tr.trackBuffer = nil
	tr.priority = 0
	tr.order = 0

	return nil
}

func (tr *trackBufferReader) listenSequences() {
	for seq := range tr.sequenceCh {
		heap.Push(tr.sequenceHeap, seq)
		select {
		case tr.ch <- struct{}{}:
		default:
		}
	}
}
