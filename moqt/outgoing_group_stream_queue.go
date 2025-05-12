package moqt

import (
	"container/heap"
	"sync"
)

func newOutgoingGroupStreamQueue(config *SubscribeConfig, scheduler *scheduler) *outgoingGroupStreamQueue {
	q := &outgoingGroupStreamQueue{
		queue:     make([]*sendGroupStream, 0, 1<<4), // TODO: Tune the initial capacity
		config:    config,
		heap:      newGroupSequenceHeap(config.GroupOrder, config.MinGroupSequence, config.MaxGroupSequence),
		scheduler: scheduler,
	}

	return q
}

// outgoingGroupStreamQueue represents a queue for outgoing group streams.
type outgoingGroupStreamQueue struct {
	queue []*sendGroupStream
	heap  *groupSequenceHeap

	id     SubscribeID
	config *SubscribeConfig
	mu     sync.Mutex

	scheduler *scheduler
}

// Enqueue adds a new stream to the queue and maintains heap property
func (q *outgoingGroupStreamQueue) Enqueue(stream *sendGroupStream) {
	if stream == nil {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	seq := stream.GroupSequence()

	q.queue = append(q.queue, stream)

	entry := struct {
		seq   GroupSequence
		index int
	}{
		seq:   seq,
		index: len(q.queue) - 1,
	}

	//
	go func() {
		for {
			select {
			case <-stream.ch:
				q.mu.Lock()
				heap.Push(q.heap, entry)
				q.mu.Unlock()

				q.scheduler.Enqueue(q.id, q.config.TrackPriority)
			default:
				continue
			}
		}
	}()
}

func (q *outgoingGroupStreamQueue) Flush() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return
	}

	for {
		if q.heap.Len() == 0 {
			return
		}

		entry := heap.Pop(q.heap).(struct {
			seq   GroupSequence
			index int
		})

		stream := q.queue[entry.seq]

		if stream == nil {
			continue
		}

		err := stream.flush()
		if err != nil {
			return
		}
	}
}
