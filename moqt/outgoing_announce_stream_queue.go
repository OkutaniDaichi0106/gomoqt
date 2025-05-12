package moqt

import "sync"

func newOutgoingAnnounceStreamQueue() *outgoingAnnounceStreamQueue {
	return &outgoingAnnounceStreamQueue{
		queue: make([]*sendAnnounceStream, 0),
	}
}

type outgoingAnnounceStreamQueue struct {
	queue []*sendAnnounceStream
	mu    sync.Mutex
}

func (q *outgoingAnnounceStreamQueue) Enqueue(stream *sendAnnounceStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, stream)

}

func (q *outgoingAnnounceStreamQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, stream := range q.queue {
		stream.Close()
	}
}

func (q *outgoingAnnounceStreamQueue) ClearWithError(err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, stream := range q.queue {
		stream.CloseWithError(err)
	}
}
