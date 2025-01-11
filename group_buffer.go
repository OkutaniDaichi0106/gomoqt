package moqt

import (
	"errors"
	"io"
	"sync"
)

func NewGroupBuffer(seq GroupSequence, priority GroupPriority) GroupBuffer {
	return GroupBuffer{
		groupSequence: seq,
		groupPriority: priority,
		frames:        make([][]byte, 0),
		cond:          sync.NewCond(&sync.Mutex{}),
	}
}

func (g *GroupBuffer) GroupSequence() GroupSequence {
	return g.groupSequence
}

func (g *GroupBuffer) GroupPriority() GroupPriority {
	return g.groupPriority
}

type GroupBuffer struct {
	groupSequence GroupSequence
	groupPriority GroupPriority
	frames        [][]byte
	cond          *sync.Cond
	locked        bool
	pos           int
}

func (r *GroupBuffer) Read(buf []byte) (int, error) {
	frame, err := r.NextFrame()
	if err != nil {
		return 0, err
	}

	n := copy(buf, frame)
	return n, nil
}

func (r *GroupBuffer) NextFrame() ([]byte, error) {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	for r.pos >= len(r.frames) {
		if r.locked {
			return nil, io.EOF
		}
		r.cond.Wait()
	}

	frame := r.frames[r.pos]
	r.pos++
	return frame, nil
}

func (w *GroupBuffer) Write(frame []byte) (int, error) {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	if w.locked {
		return 0, errors.New("group is closed")
	}

	w.frames = append(w.frames, frame)
	w.cond.Signal()
	return len(frame), nil
}

func (w *GroupBuffer) Close() error {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	w.locked = true
	w.cond.Broadcast()
	return nil
}

type GroupReader interface {
	Group
	Read([]byte) (int, error)
	NextFrame() ([]byte, error)
}

var _ GroupReader = (*GroupBuffer)(nil)

type GroupWriter interface {
	Group
	Write([]byte) (int, error)
	Close() error
}

var _ GroupWriter = (*GroupBuffer)(nil)

func NewGroupReader(buf GroupBuffer) GroupReader {
	return &groupOnlyReader{
		GroupBuffer: buf,
	}
}

type groupOnlyReader struct {
	GroupBuffer
	pos int
}

func (r *groupOnlyReader) Read(buf []byte) (int, error) {
	frame, err := r.NextFrame()
	if err != nil {
		return 0, err
	}

	n := copy(buf, frame)
	return n, nil
}

func (r *groupOnlyReader) NextFrame() ([]byte, error) {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	for r.pos >= len(r.frames) {
		if r.locked {
			return nil, io.EOF
		}
		r.cond.Wait()
	}

	frame := r.frames[r.pos]
	r.pos++
	return frame, nil
}
