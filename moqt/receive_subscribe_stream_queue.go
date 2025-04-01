package moqt

import "sync"

func newReceiveSubscribeStreamQueue() *receiveSubscribeStreamQueue {
	return &receiveSubscribeStreamQueue{
		queue: make([]*receiveSubscribeStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type receiveSubscribeStreamQueue struct {
	queue []*receiveSubscribeStream
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *receiveSubscribeStreamQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *receiveSubscribeStreamQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *receiveSubscribeStreamQueue) Enqueue(rss *receiveSubscribeStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, rss)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receiveSubscribeStreamQueue) Dequeue() *receiveSubscribeStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) <= 0 {
		return nil
	}

	next := q.queue[0]
	q.queue = q.queue[1:]

	return next
}
