package moqt

import (
	"context"
	"sync"
)

func newSendInfoStreamQueue() *incomingInfoStreamQueue {
	return &incomingInfoStreamQueue{
		queue: make([]*sendInfoStream, 0),
		ch:    make(chan struct{}),
	}
}

type incomingInfoStreamQueue struct {
	queue []*sendInfoStream
	mu    sync.Mutex
	ch    chan struct{}
	pos   int
}

// func (q *incomingInfoStreamQueue) Len() int {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	return len(q.queue)
// }

func (q *incomingInfoStreamQueue) Enqueue(req *sendInfoStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, req)
}

// func (q *incomingInfoStreamQueue) Dequeue() *sendInfoStream {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	if len(q.queue) == q.pos {
// 		return nil
// 	}

// 	req := q.queue[q.pos]

// 	q.pos++

// 	return req
// }

// func (q *incomingInfoStreamQueue) Chan() <-chan struct{} {
// 	return q.ch
// }

func (q *incomingInfoStreamQueue) Accept(ctx context.Context) (*sendInfoStream, error) {
	for {
		q.mu.Lock()
		if q.pos <= len(q.queue) {
			stream := q.queue[q.pos]
			q.pos++

			q.mu.Unlock()
			return stream, nil
		}
		q.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-q.ch:
		}
	}
}
