package moqt

import "sync"

type FrameSequence uint64

var DefaultFrameSize = 2048

var framePool = sync.Pool{
	New: func() interface{} {
		return &Frame{
			bytes: make([]byte, 0, DefaultFrameSize),
		}
	},
}

type Frame struct {
	bytes []byte
}

// NewFrame creates a new Frame with the specified bytes.
// The bytes are not copied, so the caller must not modify the bytes after calling this function.
func NewFrame(b []byte) *Frame {
	f := framePool.Get().(*Frame)
	f.bytes = b
	return f
}

// Updated CopyBytes method to return a copy of the internal slice.
func (f Frame) CopyBytes() []byte {
	b := make([]byte, len(f.bytes))
	copy(b, f.bytes)
	return b
}

func (f Frame) Size() int {
	return len(f.bytes)
}

func (f *Frame) Release() {
	f.bytes = f.bytes[:0]
	framePool.Put(f)
}
