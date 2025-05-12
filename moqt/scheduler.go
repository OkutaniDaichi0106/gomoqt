package moqt

import (
	"container/heap"
	"context"
)

func newScheduler() *scheduler {
	return &scheduler{
		heap: newTrackPriorityHeap(),
	}
}

type scheduler struct {
	heap *trackPriorityHeap
	ch   chan struct{}
}

func (s *scheduler) Len() int {
	return s.heap.Len()
}

func (s *scheduler) Enqueue(id SubscribeID, priority TrackPriority) {
	heap.Push(s.heap, struct {
		id       SubscribeID
		priority TrackPriority
	}{
		id:       id,
		priority: priority,
	})

	select {
	case s.ch <- struct{}{}:
	default:
	}
}

func (s *scheduler) Dequeue(ctx context.Context) (SubscribeID, error) {
	for {
		if s.heap.Len() > 0 {
			entry := heap.Pop(s.heap).(struct {
				id       SubscribeID
				priority TrackPriority
			})

			return entry.id, nil
		}

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-s.ch:
		}
	}
}
