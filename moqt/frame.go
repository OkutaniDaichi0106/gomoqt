package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

// NewFrame creates a new Frame with the specified bytes.
// Note: The given byte slice is referenced directly. If you modify the original slice after calling NewFrame,
// the Frame's contents will also be affected. Frame is designed to be immutable after creation.
// Returns a pointer to a Frame containing the provided payload.
func NewFrame(b []byte) *Frame {
	return &Frame{
		message: b,
	}
}

// Frame represents a data frame containing a payload.
type Frame struct {
	message message.FrameMessage
}

// Bytes returns a copy of the payload bytes contained in the Frame.
// The returned slice is a copy and can be safely modified by the caller.
func (f *Frame) Bytes() []byte {
	data := make([]byte, f.message.Len())
	copy(data, f.message)
	return data
}

// Len returns the length of the payload in bytes.
func (f *Frame) Len() int {
	return f.message.Len()
}

// Cap returns the capacity of the underlying payload slice.
func (f *Frame) Cap() int {
	return cap(f.message)
}
