package moqtrelay

import (
	"bytes"
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func NewGroupBuffer(seq moqt.GroupSequence, priority moqt.GroupPriority, buf []byte) GroupBuffer {
	return GroupBuffer{
		groupSequence: seq,
		groupPriority: priority,
		data:          bytes.NewBuffer(buf),
		cond:          sync.NewCond(&sync.Mutex{}),
	}
}

func (g *GroupBuffer) GroupSequence() moqt.GroupSequence {
	return g.groupSequence
}

func (g *GroupBuffer) GroupPriority() moqt.GroupPriority {
	return g.groupPriority
}

type GroupBuffer struct {
	groupSequence moqt.GroupSequence
	groupPriority moqt.GroupPriority
	data          *bytes.Buffer
	cond          *sync.Cond
	closed        bool
}

func (r *GroupBuffer) Read(buf []byte) (int, error) {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	if r.closed {
		return 0, errors.New("group is closed")
	}

	return r.data.Read(buf)
}

func (r *GroupBuffer) ReadFrame() ([]byte, error) {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	if r.closed {
		return nil, errors.New("group is closed")
	}

	var fm message.FrameMessage
	err := fm.Decode(r.data)
	if err != nil {
		return nil, err
	}

	return fm.Payload, nil
}

func (w *GroupBuffer) Write(buf []byte) (int, error) {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	if w.closed {
		return 0, errors.New("group is closed")
	}

	n, err := w.data.Write(buf)
	if err != nil {
		return n, err
	}

	w.cond.Signal()

	return n, err
}

func (w *GroupBuffer) WriteFrame(frame []byte) error {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	if w.closed {
		return errors.New("group is closed")
	}

	fm := message.FrameMessage{
		Payload: frame,
	}
	err := fm.Encode(w.data)
	if err != nil {
		return err
	}

	w.cond.Signal()

	return nil
}

func (w *GroupBuffer) Close() error {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	w.closed = true
	w.cond.Broadcast()
	return nil
}

var _ moqt.GroupReader = (*GroupBuffer)(nil)

var _ moqt.GroupWriter = (*GroupBuffer)(nil)

func NewGroupReader(buf GroupBuffer) moqt.GroupReader {
	return &groupReader{
		GroupBuffer: buf,
	}
}

var _ moqt.GroupReader = (*groupReader)(nil)

type groupReader struct {
	GroupBuffer

	off int
}

func (r *groupReader) Read(buf []byte) (int, error) {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	if r.closed {
		return 0, errors.New("group is closed")
	}

	n, err := r.data.Read(buf)
	r.off += n

	return n, err
}

func (r *groupReader) ReadFrame() ([]byte, error) {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	if r.closed {
		return nil, errors.New("group is closed")
	}

	var fm message.FrameMessage
	err := fm.Decode(r.data)
	if err != nil {
		return nil, err
	}

	return fm.Payload, nil
}
