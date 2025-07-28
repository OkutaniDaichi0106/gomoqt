package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

// NewFrame creates a new Frame with the specified bytes.
// Note: The given byte slice is referenced directly. If you modify the original slice after calling NewFrame,
// the Frame's contents will also be affected. Frame is designed to be immutable after creation.
func NewFrame(b []byte) *Frame {
	return &Frame{
		message: &message.FrameMessage{
			Payload: b,
		},
	}
}

type Frame struct {
	message *message.FrameMessage
}

func (f *Frame) Bytes() []byte {
	data := make([]byte, f.message.Len())
	copy(data, f.message.Payload)
	return data
}

func (f *Frame) Len() int {
	return f.message.Len()
}

func (f *Frame) Cap() int {
	return cap(f.message.Payload)
}
