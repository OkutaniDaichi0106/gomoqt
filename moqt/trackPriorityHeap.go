package moqt

import "container/heap"

func newTrackPriorityHeap() *trackPriorityHeap {
	h := &trackPriorityHeap{
		queue: make([]struct {
			id       SubscribeID
			priority TrackPriority
		}, 0),
	}

	heap.Init(h)

	return h
}

type trackPriorityHeap struct {
	queue []struct {
		id       SubscribeID
		priority TrackPriority
	}
}

func (h *trackPriorityHeap) Len() int {
	return len(h.queue)
}

func (h *trackPriorityHeap) Less(i, j int) bool {
	return h.queue[i].priority < h.queue[j].priority
}

func (h *trackPriorityHeap) Swap(i, j int) {
	h.queue[i], h.queue[j] = h.queue[j], h.queue[i]
}

func (h *trackPriorityHeap) Push(x interface{}) {
	h.queue = append(h.queue, x.(struct {
		id       SubscribeID
		priority TrackPriority
	}))
}

func (h *trackPriorityHeap) Pop() interface{} {
	old := h.queue
	n := len(old)
	x := old[n-1]
	h.queue = old[0 : n-1]
	return x
}
