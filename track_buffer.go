package moqt

import (
	"sort"
	"sync"
)

func NewTrackBuffer() *TrackBuffer {
	return &TrackBuffer{
		groupSeqs: make(map[GroupSequence]struct{}),
		groupBufs: make([]GroupBuffer, 0),
	}
}

type TrackBuffer struct {
	groupSeqs map[GroupSequence]struct{}
	groupBufs []GroupBuffer
	mu        sync.Mutex
}

func (t *TrackBuffer) AddGroup(g GroupBuffer) error {
	if _, ok := t.groupSeqs[g.GroupSequence()]; ok {
		return ErrDuplicatedGroup
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Add the sequence
	t.groupSeqs[g.GroupSequence()] = struct{}{}

	// Add the group
	t.groupBufs = append(t.groupBufs, g)

	// Sort the groups by sequence
	sort.Slice(t.groupBufs, func(i, j int) bool {
		return t.groupBufs[i].GroupSequence() < t.groupBufs[j].GroupSequence()
	})

	return nil
}
