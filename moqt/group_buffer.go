package moqt

import (
	"sync"
	"time"
)

// NewGroupBuffer creates a new GroupBuffer with the specified group sequence and initial capacity.
func newGroupBuffer(seq GroupSequence, size int) *GroupBuffer {
	if size <= 0 {
		size = DefaultGroupBufferSize
	}
	return &GroupBuffer{
		groupSequence: seq,
		cond:          sync.NewCond(&sync.Mutex{}),
		frames:        make([]Frame, 0, size),
	}
}

type GroupBuffer struct {
	groupSequence GroupSequence

	// frames
	frames []Frame

	closed bool

	closedErr error

	cond *sync.Cond

	deadline      *time.Time
	deadlineTimer *time.Timer
}

func (g *GroupBuffer) GroupSequence() GroupSequence {
	return g.groupSequence
}

func (g *GroupBuffer) SetDeadline(t time.Time) {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if g.deadlineTimer != nil {
		g.deadlineTimer.Stop()
	}

	g.deadline = &t
}

// Release resets the buffer with a new group sequence.
func (g *GroupBuffer) Release() {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if len(g.frames) > 0 {
		g.frames = g.frames[:0]
	}

	g.closed = false
	g.cond = nil
	g.deadline = nil
	g.deadlineTimer = nil
	// TODO: Release to sync.Pool
}

var DefaultGroupBufferSize = defaultBufferSize // TODO:

const defaultBufferSize = 1024 * 1024 // 1MB
