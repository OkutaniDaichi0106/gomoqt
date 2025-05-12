package moqt

import "sync"

func newOutgoingSubscribeStreamQueue() *outgoingSubscribeStreamQueue {
	return &outgoingSubscribeStreamQueue{
		queue: make([]*sendSubscribeStream, 0),
	}
}

type outgoingSubscribeStreamQueue struct {
	queue []*sendSubscribeStream
	mu    sync.Mutex
}

// func (q *outgoingSubscribeStreamQueue) Len() int {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()
// 	return len(q.queue)
// }

// func (q *outgoingSubscribeStreamQueue) Chan() <-chan struct{} {
// 	return q.ch
// }

func (q *outgoingSubscribeStreamQueue) Enqueue(stream *sendSubscribeStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, stream)
}

func (q *outgoingSubscribeStreamQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, stream := range q.queue {
		stream.Close()
	}
}

func (q *outgoingSubscribeStreamQueue) ClearWithError(err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, stream := range q.queue {
		stream.CloseWithError(err)
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
