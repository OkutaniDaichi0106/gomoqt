package gomoq

// import (
// 	"container/heap"
// 	"fmt"
// )

// // Publisher Priority Control
// type PublisherPriorityQueue []*Object

// func (ppq PublisherPriorityQueue) Len() int {
// 	return len(ppq)
// }

// func (ppq PublisherPriorityQueue) Less(i, j int) bool {
// 	return ppq[i].PublisherPriority < ppq[j].PublisherPriority
// }

// func (ppq PublisherPriorityQueue) Swap(i, j int) {
// 	ppq[i], ppq[j] = ppq[j], ppq[i]
// 	// ppq[i]
// 	// ppq[j].index
// }

// func (ppq *PublisherPriorityQueue) Push(x any) {
// 	message := x.(*Object)
// 	*ppq = append(*ppq, message)

// }

// func (ppq *PublisherPriorityQueue) Pop() any {
// 	old := *ppq
// 	n := len(old)
// 	x := old[n-1]
// 	*ppq = old[0 : n-1]
// 	return x
// }

// func Example() {
// 	ppq := make(PublisherPriorityQueue, 0)
// 	heap.Init(&ppq)

// 	//
// 	heap.Push(&ppq, &Object{})
// 	heap.Push(&ppq, &Object{})
// 	heap.Push(&ppq, &Object{})

// 	//
// 	for ppq.Len() > 0 {
// 		message := heap.Pop(&ppq).(*Object)
// 		fmt.Print(message)
// 	}

// }
