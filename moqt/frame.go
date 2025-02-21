package moqt

type FrameSequence uint64

type Frame struct {
	bytes []byte
}

// NewFrame creates a new Frame with the specified bytes.
// The bytes are copied, so the caller can modify the bytes after calling this function.
func NewFrame(b []byte) *Frame {
	newBytes := make([]byte, len(b))
	copy(newBytes, b)
	return newFrame(newBytes)
}

// newFrame creates a new Frame with the specified bytes.
// The bytes are not copied, so the caller must not modify the bytes after calling this function.
func newFrame(b []byte) *Frame {
	return &Frame{
		bytes: b,
	}
}
