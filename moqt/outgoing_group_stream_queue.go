package moqt

import (
	"container/heap"
	"sync"
)

func newOutgoingGroupStreamQueue(config func() *SubscribeConfig, order GroupOrder) *outgoingGroupStreamQueue {
	q := &outgoingGroupStreamQueue{
		queue:  make([]*sendGroupStream, 0, 1<<4), // TODO: Tune the initial capacity
		config: config,
		heap:   newGroupSequenceHeap(order),
	}

	return q
}

// outgoingGroupStreamQueue represents a queue for outgoing group streams.
type outgoingGroupStreamQueue struct {
	mu    sync.Mutex
	queue []*sendGroupStream
	heap  *groupSequenceHeap

	config func() *SubscribeConfig
}

// Enqueue adds a new stream to the queue and maintains heap property
func (q *outgoingGroupStreamQueue) enqueue(stream *sendGroupStream) {
	if stream == nil {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, stream)

	entry := struct {
		seq   GroupSequence
		index int
	}{
		seq:   stream.GroupSequence(),
		index: len(q.queue) - 1,
	}

	heap.Push(q.heap, entry)
}

func (q *outgoingGroupStreamQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, stream := range q.queue {
		stream.Close()
	}
}
