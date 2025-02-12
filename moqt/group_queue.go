package moqt

import "container/heap"

func newGroupBufferHeap(groupOrder GroupOrder) *groupBufferHeap {
	buf := &groupBufferHeap{
		queue:      make([]*GroupBuffer, 0), // TODO: Tune the initial capacity
		groupOrder: groupOrder,
	}

	heap.Init(buf)

	return buf
}

type groupBufferHeap struct {
	queue      []*GroupBuffer
	groupOrder GroupOrder
}

func (q *groupBufferHeap) Len() int {
	return len(q.queue)
}

func (q *groupBufferHeap) Push(x interface{}) {
	gb, ok := x.(*GroupBuffer)
	if !ok {
		// Type mismatch handling (omitted)
		return
	}
	// Append the new element
	q.queue = append(q.queue, gb)
}

func (q *groupBufferHeap) Less(i, j int) bool {
	switch q.groupOrder {
	case ASCENDING:
		return q.queue[i].GroupSequence() < q.queue[j].GroupSequence()
	case DESCENDING:
		return q.queue[i].GroupSequence() > q.queue[j].GroupSequence()
	case DEFAULT:
		return true
	default:
		return false
	}
}

func (q *groupBufferHeap) Swap(i, j int) {
	q.queue[i], q.queue[j] = q.queue[j], q.queue[i]
}

func (q *groupBufferHeap) Pop() interface{} {
	old := q.queue
	n := len(old) - 1
	gb := old[n]
	q.queue = old[:n]
	return gb

}
