package moqt

import (
	"sync"
)

func newOutgoingGroupStreamQueue() *outgoingGroupStreamQueue {
	q := &outgoingGroupStreamQueue{
		queue: make(map[*sendGroupStream]struct{}), // TODO: Tune the initial capacity
	}

	return q
}

// outgoingGroupStreamQueue represents a queue for outgoing group streams.
type outgoingGroupStreamQueue struct {
	mu    sync.Mutex
	queue map[*sendGroupStream]struct{}
}

func (q *outgoingGroupStreamQueue) add(str *sendGroupStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue[str] = struct{}{}
}

func (q *outgoingGroupStreamQueue) remove(str *sendGroupStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.queue, str)
}

func (q *outgoingGroupStreamQueue) clear(reason error) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for stream := range q.queue {
		stream.CloseWithError(reason)
	}

	return nil
}
