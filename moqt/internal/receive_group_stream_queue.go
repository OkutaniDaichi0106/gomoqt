package internal

import (
	"container/heap"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

func newGroupReceiverQueue(sm message.SubscribeMessage) *receiveGroupStreamQueue {
	q := &receiveGroupStreamQueue{
		queue: make([]*ReceiveGroupStream, 0, 1<<4), // TODO: Tune the initial capacity
		ch:    make(chan struct{}, 1),
		sm:    sm,
	}

	heap.Init(q)
	return q
}

type receiveGroupStreamQueue struct {
	queue []*ReceiveGroupStream
	ch    chan struct{}
	mu    sync.Mutex
	sm    message.SubscribeMessage
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

	switch q.sm.GroupOrder {
	case message.GroupOrderDefault:
		return true
	case message.GroupOrderAscending:
		return q.queue[i].GroupMessage.GroupSequence < q.queue[j].GroupMessage.GroupSequence
	case message.GroupOrderDescending:
		return q.queue[i].GroupMessage.GroupSequence > q.queue[j].GroupMessage.GroupSequence
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

	stream, ok := x.(*ReceiveGroupStream)
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
func (q *receiveGroupStreamQueue) Enqueue(stream *ReceiveGroupStream) error {
	if stream == nil {
		return ErrInternalError
	}

	if stream.GroupMessage.GroupSequence < q.sm.MinGroupSequence || stream.GroupMessage.GroupSequence > q.sm.MaxGroupSequence {
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
func (q *receiveGroupStreamQueue) Dequeue() *ReceiveGroupStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	x := heap.Pop(q)
	stream, ok := x.(*ReceiveGroupStream)
	if !ok {
		return nil
	}
	return stream

}
