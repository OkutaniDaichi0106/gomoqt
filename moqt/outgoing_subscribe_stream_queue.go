package moqt

import "sync"

func newOutgoingSubscribeStreamQueue() *outgoingSubscribeStreamQueue {
	return &outgoingSubscribeStreamQueue{
		queue: make([]*SendSubscribeStream, 0),
	}
}

type outgoingSubscribeStreamQueue struct {
	queue []*SendSubscribeStream
	mu    sync.Mutex
}

func (q *outgoingSubscribeStreamQueue) enqueue(stream *SendSubscribeStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, stream)
}

func (q *outgoingSubscribeStreamQueue) close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, stream := range q.queue {
		stream.close()
	}
}

func (q *outgoingSubscribeStreamQueue) closeWithError(err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, stream := range q.queue {
		stream.closeWithError(err)
	}
}

//

// func (q *outgoingSubscribeStreamQueue) Dequeue() *sendSubscribeStream {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	if len(q.queue) <= q.pos {
// 		return nil
// 	}

// 	next := q.queue[q.pos]
// 	q.pos++

// 	return next
// }
