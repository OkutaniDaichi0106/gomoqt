package moqt

import (
	"container/heap"
	"context"
	"sync"
)

func newGroupReceiverQueue(id SubscribeID, path TrackPath, config *SubscribeConfig) *incomingGroupStreamQueue {
	q := &incomingGroupStreamQueue{
		queue:  make([]*receiveGroupStream, 0, 1<<4),
		heap:   newGroupSequenceHeap(config.GroupOrder),
		ch:     make(chan struct{}, 1),
		id:     id,
		path:   path,
		config: config,
	}

	return q
}

var _ TrackReader = (*incomingGroupStreamQueue)(nil)

type incomingGroupStreamQueue struct {
	queue []*receiveGroupStream
	heap  *groupSequenceHeap
	ch    chan struct{}
	mu    sync.Mutex
	// subscription SentSubscription
}

// // Len implements heap.Interface
// func (q *incomingGroupStreamQueue) Len() int {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()
// 	return len(q.queue)
// }

// // Less implements heap.Interface
// func (q *incomingGroupStreamQueue) Less(i, j int) bool {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	if i >= len(q.queue) || j >= len(q.queue) {
// 		return false
// 	}

// 	switch q.config.GroupOrder {
// 	case GroupOrderDefault:
// 		return true
// 	case GroupOrderAscending:
// 		return q.queue[i].GroupSequence() < q.queue[j].GroupSequence()
// 	case GroupOrderDescending:
// 		return q.queue[i].GroupSequence() > q.queue[j].GroupSequence()
// 	default:
// 		return false
// 	}
// }

// Swap implements heap.Interface
// func (q *incomingGroupStreamQueue) Swap(i, j int) {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	if i >= len(q.queue) || j >= len(q.queue) {
// 		return
// 	}
// 	q.queue[i], q.queue[j] = q.queue[j], q.queue[i]
// }

// // Push implements heap.Interface
// func (q *incomingGroupStreamQueue) Push(x interface{}) {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	stream, ok := x.(*receiveGroupStream)
// 	if !ok || stream == nil {
// 		return
// 	}
// 	q.queue = append(q.queue, stream)
// }

// // Pop implements heap.Interface
// func (q *incomingGroupStreamQueue) Pop() interface{} {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	if len(q.queue) == 0 {
// 		return nil
// 	}

// 	n := len(q.queue) - 1
// 	item := q.queue[n]
// 	q.queue = q.queue[:n]
// 	return item
// }

// // Chan returns the notification channel
// func (q *incomingGroupStreamQueue) Chan() <-chan struct{} {
// 	return q.ch
// }

// Enqueue adds a new stream to the queue and maintains heap property
func (q *incomingGroupStreamQueue) enqueue(stream *receiveGroupStream) {
	if stream == nil {
		return
	}

	seq := stream.GroupSequence()

	if seq < q.config.MinGroupSequence || seq > q.config.MaxGroupSequence {
		return ErrInvalidRange
	}

	q.queue = append(q.queue, stream)

	entry := struct {
		seq   GroupSequence
		index int
	}{
		seq:   seq,
		index: len(q.queue) - 1,
	}

	q.mu.Lock()
	heap.Push(q.heap, entry)
	q.mu.Unlock()

	// Send a notification (non-blocking)
	select {
	case q.ch <- struct{}{}:
	default:
	}

}

// // Dequeue removes and returns the highest priority stream
// func (q *incomingGroupStreamQueue) Dequeue() *receiveGroupStream {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	if len(q.queue) == 0 {
// 		return nil
// 	}

// 	x := heap.Pop(q)
// 	stream, ok := x.(*receiveGroupStream)
// 	if !ok {
// 		return nil
// 	}
// 	return stream
// }

func (q *incomingGroupStreamQueue) AcceptGroup(ctx context.Context) (GroupReader, error) {
	for {
		q.mu.Lock()
		if q.heap.Len() > 0 {
			seq := heap.Pop(q.heap).(GroupSequence)
			stream := q.queue[seq]

			if stream == nil {
				q.mu.Unlock()
				continue
			}

			if !q.config.IsInRange(seq) {
				stream.CancelRead(ErrGroupOutOfRange)
				q.mu.Unlock()
				continue
			}

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

func (q *incomingGroupStreamQueue) CancelRead(err error) {

}

func (q *incomingGroupStreamQueue) removeGroups(min, max GroupSequence) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.heap.ResetOrder(config.GroupOrder)
	q.config = config
}
