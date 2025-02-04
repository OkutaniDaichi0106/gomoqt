package moqtrelay

// import (
// 	"container/heap"
// 	"sync"
// )

// // groupHeap implements heap.Interface
// type groupHeap struct {
// 	bufs  []*groupBuffer
// 	order GroupOrder
// 	cond  *sync.Cond
// }

// func newGroupHeap(order GroupOrder) *groupHeap {
// 	q := &groupHeap{
// 		bufs:  make([]*groupBuffer, 0),
// 		order: order,
// 		cond:  sync.NewCond(&sync.Mutex{}),
// 	}
// 	heap.Init(q)
// 	return q
// }

// // Len implements sort.Interface
// func (q *groupHeap) Len() int {
// 	return len(q.bufs)
// }

// // Less implements sort.Interface
// // Groups are ordered by their sequence number
// func (q *groupHeap) Less(i, j int) bool {
// 	switch q.order {
// 	case DEFAULT:
// 		return true
// 	case ASCENDING:
// 		return q.bufs[i].groupSequence < q.bufs[j].groupSequence
// 	case DESCENDING:
// 		return q.bufs[i].groupSequence > q.bufs[j].groupSequence
// 	default:
// 		return false
// 	}
// }

// // Swap implements sort.Interface
// func (q *groupHeap) Swap(i, j int) {
// 	q.bufs[i], q.bufs[j] = q.bufs[j], q.bufs[i]
// }

// // Push implements heap.Interface
// func (q *groupHeap) Push(x interface{}) {
// 	group := x.(*groupBuffer)
// 	q.bufs = append(q.bufs, group)
// }

// // Pop implements heap.Interface
// func (q *groupHeap) Pop() interface{} {
// 	old := q.bufs
// 	n := len(old)
// 	if n == 0 {
// 		return nil
// 	}
// 	item := old[n-1]
// 	q.bufs = old[0 : n-1]
// 	return item
// }

// // PushGroup adds a group to the queue
// func (q *groupHeap) PushGroup(group *groupBuffer) {
// 	heap.Push(q, group)
// }

// // PopGroup removes and returns the group with the lowest sequence number
// func (q *groupHeap) PopGroup() *groupBuffer {
// 	if q.Len() == 0 {
// 		return nil
// 	}
// 	return heap.Pop(q).(*groupBuffer)
// }

// // Peek returns the group with the lowest sequence number without removing it
// func (q *groupHeap) Peek() *groupBuffer {
// 	if q.Len() == 0 {
// 		return nil
// 	}
// 	return q.bufs[0]
// }
