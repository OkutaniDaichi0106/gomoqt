package message

import (
	"io"
	"sync"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * Frame Message {
 *   Message Length (varint),
 *   Payload ([]byte),
 * }
 */

func NewFrameMessage(payload []byte) *FrameMessage {
	fm := framePool.Get().(*FrameMessage)
	if cap(fm.Payload) < len(payload) {
		fm.Payload = make([]byte, len(payload))
	} else {
		fm.Payload = fm.Payload[:len(payload)]
	}
	copy(fm.Payload, payload)
	return fm
}

type FrameMessage struct {
	Payload []byte
}

func (fm *FrameMessage) Len() int {
	return bytesLen(fm.Payload)
}

func (fm *FrameMessage) Encode(w io.Writer) (int, error) {
	b := getBytes()
	defer putBytes(b)

	b = AppendBytes(b, fm.Payload)

	return w.Write(b)
}

func (fm *FrameMessage) Decode(r io.Reader) (int, error) {
	var err error
	var n int

	fm.Payload, n, err = ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	return n, nil
}

var framePool = sync.Pool{
	New: func() any {
		return &FrameMessage{
			Payload: getBytes(),
		}
	},
}

var DefaultFrameSize = 2048

// CopyBytes method returns a copy of the internal slice.
func (f *FrameMessage) CopyBytes() []byte {
	b := make([]byte, len(f.Payload))
	copy(b, f.Payload)
	return b
}

func (f FrameMessage) Size() int {
	return len(f.Payload)
}

func (f *FrameMessage) Release() {
	f.Payload = f.Payload[:0]
	framePool.Put(f)
}
