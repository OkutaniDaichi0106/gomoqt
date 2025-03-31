package moqt

import "github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"

type FrameSequence uint64

var DefaultFrameSize = 2048

type Frame interface {
	// CopyBytes returns a copy of the internal slice.
	CopyBytes() []byte

	// Size returns the size of the internal slice.
	Size() int

	// Release releases the frame back to the pool.
	Release()
}

// NewFrame creates a new Frame with the specified bytes.
func NewFrame(b []byte) Frame {
	return message.NewFrameMessage(b)
}
