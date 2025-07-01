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

func (fm *FrameMessage) Encode(w io.Writer) error {
	b := getBytes()
	defer putBytes(b)

	b = AppendBytes(b, fm.Payload)

	_, err := w.Write(b)
	return err
}

func (fm *FrameMessage) Decode(r io.Reader) error {
	var err error

	fm.Payload, _, err = ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		return err
	}

	return nil
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
