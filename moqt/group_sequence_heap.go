package moqt

import (
	"container/heap"
)

func newGroupSequenceHeap(order GroupOrder) *groupSequenceHeap {
	h := &groupSequenceHeap{
		queue: make([]struct {
			seq   GroupSequence
			index int
		}, 0), // TODO: Tune the initial capacity
		order: order,
	}

	heap.Init(h)

	return h
}

var _ heap.Interface = (*groupSequenceHeap)(nil)

type groupSequenceHeap struct {
	queue []struct {
		seq   GroupSequence
		index int
	}
	order GroupOrder
}

func (q *groupSequenceHeap) Len() int {
	return len(q.queue)
}

func (q *groupSequenceHeap) Push(x any) {

	entry, ok := x.(struct {
		seq   GroupSequence
		index int
	})
	if !ok {
		// Type mismatch handling (omitted)
		return
	}

	// Append the new element
	q.queue = append(q.queue, entry)
}

func (q *groupSequenceHeap) Less(i, j int) bool {
	switch q.order {
	case GroupOrderAscending:
		return q.queue[i].seq < q.queue[j].seq
	case GroupOrderDescending:
		return q.queue[i].seq > q.queue[j].seq
	case GroupOrderDefault:
		return true
	default:
		return false
	}
}

func (q *groupSequenceHeap) Swap(i, j int) {
	q.queue[i], q.queue[j] = q.queue[j], q.queue[i]
}

func (q *groupSequenceHeap) Pop() any {
	old := q.queue
	n := len(old)
	if n == 0 {
		// Return zero value when popping from empty heap
		return struct {
			seq   GroupSequence
			index int
		}{}
	}
	gb := old[n-1]
	q.queue = old[:n-1]
	return gb
}

func (q *groupSequenceHeap) resetOrder(order GroupOrder) {
	q.order = order
}
