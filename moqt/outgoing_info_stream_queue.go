package moqt

import "sync"

func newOutgoingInfoStreamQueue() *outgoingInfoStreamQueue {
	return &outgoingInfoStreamQueue{
		queue: make([]*sendInfoStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type outgoingInfoStreamQueue struct {
	queue []*sendInfoStream
	mu    sync.Mutex
	ch    chan struct{}
	pos   int
}

func (q *outgoingInfoStreamQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queue)
}

func (q *outgoingInfoStreamQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *outgoingInfoStreamQueue) Enqueue(stream *sendInfoStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, stream)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *outgoingInfoStreamQueue) Dequeue() *sendInfoStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) <= q.pos {
		return nil
	}

	next := q.queue[q.pos]
	q.pos++

	return next
}
