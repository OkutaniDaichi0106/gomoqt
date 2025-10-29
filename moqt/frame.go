package moqt

import (
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
)

// Frame represents a MOQ frame with optimized encoding for MOQT protocol.
//
// The Frame is designed for efficient encoding/decoding of MOQ payloads:
// - buf: 8-byte buffer combining header (length varint) and payload data
// - header: 8-byte array used to store the length varint without allocations
// - body: slice pointing into buf[8:] containing the actual frame payload
//
// This design avoids creating additional byte buffers during encode operations.
// The length is encoded directly into the header array, then written with the payload.
//
// Access the payload via the Body() method to maintain internal consistency.
type Frame struct {
	buf    []byte
	header [8]byte
	body   []byte
}

// NewFrame creates a new Frame with the specified payload capacity.
// The frame is initialized with empty payload and ready for data to be appended.
func NewFrame(cap int) *Frame {
	f := &Frame{}
	f.init(cap)
	return f
}

// Reset clears the frame payload while preserving the buffer capacity.
// This allows the frame to be reused without reallocation.
func (f *Frame) Reset() {
	f.body = f.body[:0]
}

// Body returns a reference to the frame payload bytes.
// The returned slice references the internal buffer and should not be modified.
// Use Write() to add data or Reset() to clear the frame.
func (f *Frame) Body() []byte {
	return f.body
}

func (f *Frame) init(cap int) {
	f.buf = make([]byte, 8+cap)
	body := f.buf[8:8]
	if f.body != nil {
		body = body[:len(f.body)]
		copy(body, f.body)
	}
	f.body = body
}

// append appends bytes to the frame payload, growing the buffer if necessary.
// If additional capacity is needed, the buffer is reallocated and data is copied.
// This is an internal method used by Write() and Clone().
func (f *Frame) append(b []byte) {
	if len(b)+len(f.body) > cap(f.body) {
		// Reallocate the body buffer if necessary
		cap := min(len(f.body)+len(b), 2*cap(f.body))
		f.init(cap)
	}

	f.body = append(f.body, b...)
}

// Len returns the current length of the payload in bytes.
func (f *Frame) Len() int {
	return len(f.body)
}

// Cap returns the current capacity of the payload buffer.
func (f *Frame) Cap() int {
	return cap(f.body)
}

// encode writes the frame in MOQ format: length varint followed by payload.
// The length is encoded into the header buffer to minimize allocations.
func (f *Frame) encode(w io.Writer) error {
	l := uint64(len(f.body))
	header, size := message.WriteVarint(f.header[:0], l)
	start := 8 - size
	copy(f.buf[start:], header)
	_, err := w.Write(f.buf[start:])
	return err
}

// decode reads a MOQ frame from the reader, updating the payload.
// The payload buffer is reused or reallocated as needed.
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

// Clone creates a deep copy of the frame, including all payload data.
// The cloned frame is completely independent from the original.
func (f *Frame) Clone() *Frame {
	clone := NewFrame(f.Cap())
	clone.append(f.Body())
	return clone
}

// WriteTo writes the payload to the writer, returning the number of bytes written.
func (f *Frame) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(f.body)
	if err != nil {
		return 0, err
	}
	return int64(n), nil
}

// Write implements io.Writer interface for frame payloads.
// It appends the provided bytes to the frame and returns the number of bytes written.
func (f *Frame) Write(p []byte) (int, error) {
	f.append(p)
	return len(p), nil
}
