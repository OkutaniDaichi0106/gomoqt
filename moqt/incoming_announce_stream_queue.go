package moqt

import (
	"context"
	"sync"
)

func newSendAnnounceStreamQueue() *incomingAnnounceStreamQueue {
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

func (q *incomingAnnounceStreamQueue) Accept(ctx context.Context) (*sendAnnounceStream, error) {
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

// func (q *incomingAnnounceStreamQueue) Len() int {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()
// 	return len(q.queue)
// }

// func (q *incomingAnnounceStreamQueue) Chan() <-chan struct{} {
// 	return q.ch
// }

func (q *incomingAnnounceStreamQueue) Enqueue(stream *sendAnnounceStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, stream)

	select {
	case q.ch <- struct{}{}:
	default:
		stream.Close()
	}
}

// func (q *incomingAnnounceStreamQueue) Dequeue() *sendAnnounceStream {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	if q.Len() <= q.pos {
// 		return nil
// 	}

// 	sas := q.queue[q.pos]

// 	q.pos++

// 	return sas
// }
