package moqt

import (
	"context"
	"sync"
)

func newIncomingAnnounceStreamQueue() *incomingAnnounceStreamQueue {
	return &incomingAnnounceStreamQueue{
		queue: make([]*sendAnnounceStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type incomingAnnounceStreamQueue struct {
	queue []*sendAnnounceStream
	mu    sync.Mutex
	ch    chan struct{}
	pos   int
}

func (q *incomingAnnounceStreamQueue) accept(ctx context.Context) (*sendAnnounceStream, error) {
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

func (q *incomingAnnounceStreamQueue) enqueue(stream *sendAnnounceStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, stream)

	select {
	case q.ch <- struct{}{}:
	default:
		stream.Close()
	}
}
