package moqt

import (
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

// NewFrame creates a new Frame with the specified bytes.
// Note: The given byte slice is referenced directly. If you modify the original slice after calling NewFrame,
// the Frame's contents will also be affected. Frame is designed to be immutable after creation.
// Returns a pointer to a Frame containing the provided payload.
func newFrame(cap int) *Frame {
	buf := make([]byte, 8+cap)
	return &Frame{
		buf:    buf,
		header: [8]byte{},
		body:   buf[8:8], // Start with zero length
	}
}

// Frame represents a data frame containing a payload.
type Frame struct {
	buf    []byte
	header [8]byte
	body   []byte
}

func (f *Frame) reset() {
	f.body = f.body[:0]
}

func (f *Frame) append(b []byte) {
	if len(b)+len(f.body) > cap(f.body) {
		// Reallocate the body buffer if necessary
		cap := min(len(f.body)+len(b), 2*cap(f.body))
		newBuf := make([]byte, 8+cap)
		body := newBuf[8:]
		body = body[:len(f.body)]
		copy(body, f.body)
		f.body = body
	}

	f.body = append(f.body, b...)
}

// Bytes returns a copy of the payload bytes contained in the Frame.
// The returned slice is a copy and can be safely modified by the caller.
func (f *Frame) Bytes() []byte {
	data := make([]byte, len(f.body))
	copy(data, f.body)
	return data
}

// Len returns the length of the payload in bytes.
func (f *Frame) Len() int {
	return len(f.body)
}

// Cap returns the capacity of the underlying payload slice.
func (f *Frame) Cap() int {
	return cap(f.body)
}

func (f *Frame) encode(w io.Writer) error {
	l := uint64(len(f.body))
	// end := 8 - message.VarintLen(l)
	header, size := message.WriteVarint(f.header[:0], l)
	start := 8 - size
	copy(f.buf[start:], header)
	_, err := w.Write(f.buf[start:])
	return err
}

func (f *Frame) decode(src io.Reader) error {
	num, err := message.ReadMessageLength(src)
	if err != nil {
		return err
	}

	// If payload length is zero, reset the slice to zero length
	if num == 0 {
		f.body = f.body[:0]
		return nil
	}

	// Ensure the payload slice has enough capacity
	if cap(f.body) < int(num) {
		f.body = make([]byte, num)
	} else {
		f.body = f.body[:num]
	}

	_, err = io.ReadFull(src, f.body)

	return err
}

func (f *Frame) Clone() *Frame {
	clone := newFrame(f.Cap())
	clone.append(f.Bytes())
	return clone
}

func NewFrameBuilder(cap int) *FrameBuilder {
	return &FrameBuilder{
		frame: newFrame(cap),
	}
}

type FrameBuilder struct {
	frame *Frame
}

func (fb *FrameBuilder) Append(b []byte) {
	fb.frame.append(b)
}

func (fb *FrameBuilder) Frame() *Frame {
	return fb.frame
}

func (fb *FrameBuilder) Reset() {
	fb.frame.reset()
}
