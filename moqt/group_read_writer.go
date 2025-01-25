package moqt

import (
	"bytes"
	"errors"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type GroupBuffer interface {
	GroupReader
	GroupWriter
	Len() int
	Cap() int
	Bytes() []byte
	Truncate() error
	Reset()
	Grow(n int)
	Closed() bool
}

var _ GroupBuffer = (*groupBuffer)(nil)

func NewGroupBuffer(seq GroupSequence, buf []byte) GroupBuffer {
	return &groupBuffer{
		groupSequence: seq,
		bytes:         buf,
	}
}

type groupBuffer struct {
	groupSequence GroupSequence
	bytes         []byte
	off           int
	w_closed      bool
}

func (g *groupBuffer) GroupSequence() GroupSequence {
	return g.groupSequence
}

func (r *groupBuffer) ReadFrame() ([]byte, error) {
	var fm message.FrameMessage
	n, err := fm.Decode(bytes.NewReader(r.bytes[r.off:]))
	if err != nil {
		return nil, err
	}

	r.off += n

	return fm.Payload, nil
}

func (w *groupBuffer) WriteFrame(frame []byte) error {
	if w.w_closed {
		return errors.New("group is closed")
	}

	fm := message.FrameMessage{
		Payload: frame,
	}
	err := fm.Encode(&w.buf)
	if err != nil {
		return err
	}

	return nil
}

func (g *groupBuffer) Reset() {
	g.buf.Reset()
}

func (g *groupBuffer) Close() error {
	g.w_closed = true
	return nil
}

func (g *groupBuffer) Closed() bool {
	return g.w_closed
}

func (g *groupBuffer) Truncate() error {
	if g.w_closed {
		return errors.New("group is closed")
	}
	g.buf.Truncate(0)
	return nil
}

func (g *groupBuffer) Grow(n int) {
	g.buf.Grow(n)
}

func (g *groupBuffer) FrameOffset(seq FrameSequence) int {

	return g.buf.Len()
}
