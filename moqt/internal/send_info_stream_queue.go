package internal

import "sync"

func newReceiveInfoStreamQueue() *sendInfoStreamQueue {
	return &sendInfoStreamQueue{
		queue: make([]*SendInfoStream, 0),
		ch:    make(chan struct{}),
	}
}

type sendInfoStreamQueue struct {
	queue []*SendInfoStream
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *sendInfoStreamQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *sendInfoStreamQueue) Enqueue(req *SendInfoStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, req)
}

func (q *sendInfoStreamQueue) Dequeue() *SendInfoStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	req := q.queue[0]
	q.queue = q.queue[1:]

	return req
}

func (q *sendInfoStreamQueue) Chan() <-chan struct{} {
	return q.ch
}
