package moqt

import (
	"context"
	"sync"
)

func newReceiveSubscribeStreamQueue() *incomingSubscribeStreamQueue {
	return &incomingSubscribeStreamQueue{
		queue: make([]*receiveSubscribeStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type incomingSubscribeStreamQueue struct {
	queue []*receiveSubscribeStream
	mu    sync.Mutex
	ch    chan struct{}
	pos   int
}

// func (q *incomingSubscribeStreamQueue) Len() int {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	return len(q.queue)
// }

// func (q *incomingSubscribeStreamQueue) Chan() <-chan struct{} {
// 	return q.ch
// }

func (q *incomingSubscribeStreamQueue) Enqueue(rss *receiveSubscribeStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, rss)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

// func (q *incomingSubscribeStreamQueue) Dequeue() *receiveSubscribeStream {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	if len(q.queue) <= q.pos {
// 		return nil
// 	}

// 	next := q.queue[q.pos]

// 	q.pos++

// 	return next
// }

func (q *incomingSubscribeStreamQueue) Accept(ctx context.Context) (*receiveSubscribeStream, error) {
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
