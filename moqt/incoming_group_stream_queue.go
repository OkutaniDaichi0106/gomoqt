package moqt

import (
	"context"
	"sync"
)

func newIncomingGroupStreamQueue(config func() *SubscribeConfig) *incomingGroupStreamQueue {
	return &incomingGroupStreamQueue{
		queue:  make([]*receiveGroupStream, 0, 1<<4),
		ch:     make(chan struct{}, 1),
		config: config,
	}
}

// var _ TrackReader = (*incomingGroupStreamQueue)(nil)

type incomingGroupStreamQueue struct {
	queue []*receiveGroupStream
	ch    chan struct{}
	mu    sync.Mutex

	dequeued map[*receiveGroupStream]struct{}

	config func() *SubscribeConfig
}

// Enqueue adds a new stream to the queue and maintains heap property
func (q *incomingGroupStreamQueue) enqueue(stream *receiveGroupStream) {
	if stream == nil {
		return
	}

	seq := stream.GroupSequence()

	if q.config().IsInRange(seq) {
		stream.CancelRead(ErrGroupOutOfRange)
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, stream)

	// Send a notification (non-blocking)
	select {
	case q.ch <- struct{}{}:
	default:
	}

}

func (q *incomingGroupStreamQueue) remove(stream *receiveGroupStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.dequeued, stream)
}

func (q *incomingGroupStreamQueue) dequeue(ctx context.Context) (*receiveGroupStream, error) {
	for {
		q.mu.Lock()
		if len(q.queue) > 0 {
			stream := q.queue[0]

			q.queue = q.queue[1:]

			if stream == nil {
				q.mu.Unlock()
				continue
			}

			if !q.config().IsInRange(stream.GroupSequence()) {
				stream.CancelRead(ErrGroupOutOfRange)
				q.mu.Unlock()
				continue
			}

			q.dequeued[stream] = struct{}{}

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

func (q *incomingGroupStreamQueue) clear(reason GroupError) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, stream := range q.queue {
		stream.CancelRead(reason)
	}

	for stream := range q.dequeued {
		stream.CancelRead(reason)
	}

	q.queue = q.queue[:0]
	q.dequeued = nil

	return reason
}
