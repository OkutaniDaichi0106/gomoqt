package internal

import "sync"

func newSendAnnounceStreamQueue() *sendAnnounceStreamQueue {
	return &sendAnnounceStreamQueue{
		queue: make([]*SendAnnounceStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type sendAnnounceStreamQueue struct {
	queue []*SendAnnounceStream
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *sendAnnounceStreamQueue) Len() int {
	return len(q.queue)
}

func (q *sendAnnounceStreamQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *sendAnnounceStreamQueue) Enqueue(interest *SendAnnounceStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, interest)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *sendAnnounceStreamQueue) Dequeue() *SendAnnounceStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.Len() <= 0 {
		return nil
	}

	sas := q.queue[0]
	q.queue = q.queue[1:]

	return sas
}
