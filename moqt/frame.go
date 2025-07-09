package moqt

import (
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

// type FrameSequence uint64

// NewFrame creates a new Frame with the specified bytes.
func NewFrame(b []byte) *Frame {
	return &Frame{message: message.NewFrameMessage(b)}
}

type Frame struct {
	message *message.FrameMessage
}

func (f *Frame) Decode(r io.Reader) error {
	return f.message.Decode(r)
}

func (f *Frame) Encode(w io.Writer) error {
	return f.message.Encode(w)
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
