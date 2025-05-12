package moqt

import "github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"

type FrameSequence uint64

var DefaultFrameSize = 2048

// type Frame interface {
// 	// CopyBytes returns a copy of the internal slice.
// 	CopyBytes() []byte

// 	// Size returns the size of the internal slice.
// 	Size() int

// 	// Release releases the frame back to the pool.
// 	// Release()
// }

// NewFrame creates a new Frame with the specified bytes.
func NewFrame(b []byte) Frame {
	return Frame{message: message.NewFrameMessage(b)}
}

type Frame struct {
	message *message.FrameMessage
}

func (f *Frame) CopyBytes() []byte {
	return f.message.CopyBytes()
}

func (f *Frame) Size() int {
	return f.message.Size()
}

func (f *Frame) Release() {
	f.message.Release()
}
