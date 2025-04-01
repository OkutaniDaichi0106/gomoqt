package moqt

import (
	"container/heap"
	"sync"
)

func newGroupReceiverQueue(id SubscribeID, path TrackPath, config SubscribeConfig) *receiveGroupStreamQueue {
	q := &receiveGroupStreamQueue{
		queue:  make([]*receiveGroupStream, 0, 1<<4), // TODO: Tune the initial capacity
		ch:     make(chan struct{}, 1),
		id:     id,
		path:   path,
		config: config,
	}

	heap.Init(q)
	return q
}

type receiveGroupStreamQueue struct {
	queue  []*receiveGroupStream
	ch     chan struct{}
	mu     sync.Mutex
	id     SubscribeID
	path   TrackPath
	config SubscribeConfig
}

// Len implements heap.Interface
func (q *receiveGroupStreamQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queue)
}

// Less implements heap.Interface
func (q *receiveGroupStreamQueue) Less(i, j int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if i >= len(q.queue) || j >= len(q.queue) {
		return false
	}

	switch q.config.GroupOrder {
	case GroupOrderDefault:
		return true
	case GroupOrderAscending:
		return q.queue[i].GroupSequence() < q.queue[j].GroupSequence()
	case GroupOrderDescending:
		return q.queue[i].GroupSequence() > q.queue[j].GroupSequence()
	default:
		return false
	}
}

// Swap implements heap.Interface
func (q *receiveGroupStreamQueue) Swap(i, j int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if i >= len(q.queue) || j >= len(q.queue) {
		return
	}
	q.queue[i], q.queue[j] = q.queue[j], q.queue[i]
}

// Push implements heap.Interface
func (q *receiveGroupStreamQueue) Push(x interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	stream, ok := x.(*receiveGroupStream)
	if !ok || stream == nil {
		return
	}
	q.queue = append(q.queue, stream)
}

// Pop implements heap.Interface
func (q *receiveGroupStreamQueue) Pop() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	n := len(q.queue) - 1
	item := q.queue[n]
	q.queue = q.queue[:n]
	return item
}

// Chan returns the notification channel
func (q *receiveGroupStreamQueue) Chan() <-chan struct{} {
	return q.ch
}

// Enqueue adds a new stream to the queue and maintains heap property
func (q *receiveGroupStreamQueue) Enqueue(stream *receiveGroupStream) error {
	if stream == nil {
		return ErrInternalError
	}

	if stream.GroupSequence() < q.config.MinGroupSequence || stream.GroupSequence() > q.config.MaxGroupSequence {
		return ErrInvalidRange
	}

	q.mu.Lock()
	heap.Push(q, stream)
	q.mu.Unlock()

	// Send a notification (non-blocking)
	select {
	case q.ch <- struct{}{}:
	default:
	}

	return nil
}

// Dequeue removes and returns the highest priority stream
func (q *receiveGroupStreamQueue) Dequeue() *receiveGroupStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	x := heap.Pop(q)
	stream, ok := x.(*receiveGroupStream)
	if !ok {
		return nil
	}
	return stream

}
