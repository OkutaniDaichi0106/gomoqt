package moqtrelay

import (
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
)

func NewGroupBuffer(seq moqt.GroupSequence, buf []byte) GroupBuffer {
	return GroupBuffer{
		GroupBuffer: moqt.NewGroupBuffer(seq, buf),
		cond:        sync.NewCond(&sync.Mutex{}),
		readers:     make(map[*groupReader]struct{}),
	}
}

type GroupBuffer struct {
	moqt.GroupBuffer
	cond    *sync.Cond
	closed  bool
	readers map[*groupReader]struct{}
}

func (gb *GroupBuffer) WriteFrame(buf []byte) error {
	gb.cond.L.Lock()
	defer gb.cond.L.Unlock()

	if gb.closed {
		return errors.New("group is closed")
	}

	err := gb.WriteFrame(buf)
	if err != nil {
		return err
	}

	gb.cond.Broadcast()
	return nil
}

func (gb *GroupBuffer) ReadFrame() ([]byte, error) {
	gb.cond.L.Lock()
	defer gb.cond.L.Unlock()

	for gb.Len() == gb.off && !gb.closed {
		gb.cond.Wait()
	}

	if gb.closed && gb.data.Len() == gb.off {
		return nil, errors.New("group is closed")
	}

	return gb.ReadFrame()
}

func (gb *GroupBuffer) Close() error {
	gb.cond.L.Lock()
	defer gb.cond.L.Unlock()

	if gb.closed {
		return errors.New("group is already closed")
	}

	gb.closed = true
	gb.cond.Broadcast()
	return nil
}

func NewGroupReader(buf GroupBuffer) moqt.GroupReader {
	gr := &groupReader{
		GroupBuffer: buf,
	}
	buf.cond.L.Lock()
	buf.readers[gr] = struct{}{}
	buf.cond.L.Unlock()
	return gr
}

var _ moqt.GroupReader = (*groupReader)(nil)

type groupReader struct {
	GroupBuffer

	off int
}

func (r *groupReader) ReadFrame() ([]byte, error) {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	for r.Len() == r.off && !r.closed {
		r.cond.Wait()
	}

	if r.closed && r.Len() == r.off {
		return nil, errors.New("group is closed")
	}

	return r.ReadFrame()
}

func (r *groupReader) Close() error {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	if r.closed {
		return errors.New("group is already closed")
	}

	r.closed = true
	delete(r.GroupBuffer.readers, r)
	r.cond.Broadcast()
	return nil
}
