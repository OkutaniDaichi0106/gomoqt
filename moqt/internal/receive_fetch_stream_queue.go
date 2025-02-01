package internal

import "sync"

func newReceivedFetchQueue() *receiveFetchStreamQueue {
	return &receiveFetchStreamQueue{
		queue: make([]*ReceiveFetchStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type receiveFetchStreamQueue struct {
	queue []*ReceiveFetchStream
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *receiveFetchStreamQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *receiveFetchStreamQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *receiveFetchStreamQueue) Enqueue(fetch *ReceiveFetchStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, fetch)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receiveFetchStreamQueue) Dequeue() *ReceiveFetchStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	next := q.queue[0]
	q.queue = q.queue[1:]

	return next
}
