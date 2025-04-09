package moqt

import (
	"fmt"
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

var _ GroupWriter = (*GroupBuffer)(nil)

type GroupBuffer struct {
	groupSequence GroupSequence

	// frames
	frames []Frame

	cond *sync.Cond

	deadline      *time.Time
	deadlineTimer *time.Timer

	closed    bool
	closedErr error
}

func (g *GroupBuffer) GroupSequence() GroupSequence {
	return g.groupSequence
}

func (g *GroupBuffer) WriteFrame(frame Frame) error {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()
	if g.closed {
		if g.closedErr != nil {
			return g.closedErr
		}
		return ErrClosedGroup
	}
	g.frames = append(g.frames, frame)

	g.cond.Broadcast()

	return nil
}

func (g *GroupBuffer) SetDeadline(t time.Time) {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if g.deadlineTimer != nil {
		g.deadlineTimer.Stop()
	}

	g.deadline = &t
}

func (g *GroupBuffer) SetWriteDeadline(t time.Time) error {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if g.closed {
		if g.closedErr != nil {
			return g.closedErr
		}
		return ErrClosedGroup
	}

	g.deadline = &t

	return nil
}

func (g *GroupBuffer) CloseWithError(err error) error {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if g.closed {
		if g.closedErr != nil {
			return fmt.Errorf("group has already closed due to: %w", g.closedErr)
		}
		return ErrClosedGroup
	}

	if err == nil {
		err = ErrInternalError
	}

	g.closed = true
	g.closedErr = err

	g.cond.Broadcast()

	return nil
}

func (g *GroupBuffer) Close() error {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if g.closed {
		if g.closedErr != nil {
			return fmt.Errorf("group has already closed due to: %w", g.closedErr)
		}
		return ErrClosedGroup
	}

	g.closed = true
	g.closedErr = nil

	g.cond.Broadcast()

	return nil
}

// Release resets the buffer with a new group sequence.
func (g *GroupBuffer) Release() {
	g.cond.L.Lock()
	defer g.cond.L.Unlock()

	if len(g.frames) > 0 {
		g.frames = g.frames[:0]
	}

	g.cond = nil
	g.deadline = nil
	g.deadlineTimer = nil
	// TODO: Release to sync.Pool
}

var DefaultGroupBufferSize = defaultBufferSize // TODO:

const defaultBufferSize = 1024 * 1024 // 1MB
