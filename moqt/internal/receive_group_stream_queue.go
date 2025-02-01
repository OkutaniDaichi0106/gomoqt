package internal

import "sync"

func newGroupReceiverQueue() *receiveGroupStreamQueue {
	return &receiveGroupStreamQueue{
		queue: make([]*ReceiveGroupStream, 0), // TODO: Tune the initial capacity
		ch:    make(chan struct{}, 1),
	}
}

type receiveGroupStreamQueue struct {
	queue []*ReceiveGroupStream
	ch    chan struct{}
	mu    sync.Mutex
}

func (q *receiveGroupStreamQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *receiveGroupStreamQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *receiveGroupStreamQueue) Enqueue(stream *ReceiveGroupStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, stream)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receiveGroupStreamQueue) Dequeue() *ReceiveGroupStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	next := q.queue[0]

	q.queue = q.queue[1:]

	return next
}
