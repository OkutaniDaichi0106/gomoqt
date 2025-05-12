package moqt

import (
	"container/heap"
)

func newGroupSequenceHeap(order GroupOrder, min, max GroupSequence) *groupSequenceHeap {
	h := &groupSequenceHeap{
		queue: make([]struct {
			seq   GroupSequence
			index int
		}, 0), // TODO: Tune the initial capacity
		order: order,
		min:   min,
		max:   max,
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
	min   GroupSequence
	max   GroupSequence
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

	//
	if q.min != LatestGroupSequence && entry.seq < q.min {
		return
	}
	if q.max != NotSpecifiedGroupSequence && entry.seq > q.max {
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
	n := len(old) - 1
	gb := old[n]
	q.queue = old[:n]
	return gb
}

func (q *groupSequenceHeap) ResetConfig(order GroupOrder, min, max GroupSequence) {
	q.order = order
	q.min = min
	q.max = max
}
