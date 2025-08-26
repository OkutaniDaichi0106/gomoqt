package moqt

import (
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

// NewFrame creates a new Frame with the specified bytes.
// Note: The given byte slice is referenced directly. If you modify the original slice after calling NewFrame,
// the Frame's contents will also be affected. Frame is designed to be immutable after creation.
// Returns a pointer to a Frame containing the provided payload.
func NewFrame(cap int) *Frame {
	buf := make([]byte, 8, 8+cap)
	return &Frame{
		buf:    buf,
		header: [8]byte(buf[:8]),
		body:   buf[8:],
	}
}

// Frame represents a data frame containing a payload.
type Frame struct {
	buf    []byte
	header [8]byte
	body   []byte
}

func (f *Frame) Reset() {
	f.body = f.body[:0]
}

func (f *Frame) Append(b []byte) {
	if len(b) > cap(f.body)-len(f.body) {
		// Reallocate the body buffer if necessary
		cap := min(len(f.body)+len(b), 2*cap(f.body))
		newBuf := make([]byte, cap)
		f.header = [8]byte(newBuf[:8])
		body := newBuf[8:]
		copy(body, f.body)
		f.body = body
	}

	f.body = append(f.body, b...)

	len := uint64(len(f.body))
	start := 8 - message.VarintLen(len)
	message.WriteVarint(f.header[start:], len)
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

func (f *Frame) Encode(w io.Writer) error {
	start := 8 - message.VarintLen(uint64(len(f.body)))
	_, err := w.Write(f.buf[start:])
	return err
}

func (f *Frame) Decode(src io.Reader) error {
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

	message.WriteVarint(f.header[:], uint64(num))

	return err
}
