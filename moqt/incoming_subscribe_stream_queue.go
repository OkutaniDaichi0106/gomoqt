package moqt

import (
	"sync"
)

func newIncomingSubscriptionQueue() *incomingSubscribeStreamQueue {
	return &incomingSubscribeStreamQueue{
		dequeued: make(map[*receiveSubscribeStream]struct{}, 0),
	}
}

type incomingSubscribeStreamQueue struct {
	// queue []*receiveSubscribeStream
	mu sync.Mutex

	dequeued map[*receiveSubscribeStream]struct{}
}

func (q *incomingSubscribeStreamQueue) enqueue(rss *receiveSubscribeStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.dequeued[rss] = struct{}{}
}

func (q *incomingSubscribeStreamQueue) remove(rss *receiveSubscribeStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.dequeued, rss)
}

func (q *incomingSubscribeStreamQueue) clear(reason error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for stream := range q.dequeued {
		stream.closeWithError(reason)
	}
}
