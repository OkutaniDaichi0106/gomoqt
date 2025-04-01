package moqt

import "container/heap"

func newGroupSequenceHeap(groupOrder GroupOrder) *groupSequenceHeap {
	buf := &groupSequenceHeap{
		queue:      make([]GroupSequence, 0), // TODO: Tune the initial capacity
		groupOrder: groupOrder,
	}

	heap.Init(buf)

	return buf
}

var _ heap.Interface = (*groupSequenceHeap)(nil)

type groupSequenceHeap struct {
	queue      []GroupSequence
	groupOrder GroupOrder
}

func (q *groupSequenceHeap) Len() int {
	return len(q.queue)
}

func (q *groupSequenceHeap) Push(x any) {
	gb, ok := x.(GroupSequence)
	if !ok {
		// Type mismatch handling (omitted)
		return
	}
	// Append the new element
	q.queue = append(q.queue, gb)
}

func (q *groupSequenceHeap) Less(i, j int) bool {
	switch q.groupOrder {
	case GroupOrderAscending:
		return q.queue[i] < q.queue[j]
	case GroupOrderDescending:
		return q.queue[i] > q.queue[j]
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
