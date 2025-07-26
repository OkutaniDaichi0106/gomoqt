package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * Frame Message {
 *   Message Length (varint),
 *   Payload ([]byte),
 * }
 */

func NewFrameMessage(payload []byte) *FrameMessage {
	p := FramePool.Get(len(payload))
	copy(p, payload)
	return &FrameMessage{
		Payload: p,
	}
}

type FrameMessage struct {
	Payload []byte
}

func (fm FrameMessage) Len() int {
	var l int

	l += quicvarint.Len(uint64(len(fm.Payload)))
	l += len(fm.Payload)

	return l
}

func (fm *FrameMessage) Encode(w io.Writer) error {
	msgLen := fm.Len()
	b := pool.Get(msgLen + quicvarint.Len(uint64(msgLen)))

	b = quicvarint.Append(b, uint64(msgLen))
	b = quicvarint.Append(b, uint64(len(fm.Payload)))
	b = append(b, fm.Payload...)

	_, err := w.Write(b)
	if err != nil {
		pool.Put(b)
		return err
	}

	pool.Put(b)
	return nil
}

func (fm *FrameMessage) Decode(src io.Reader) error {
	num, err := ReadVarint(src)
	if err != nil {
		return err
	}

	b := pool.Get(int(num))[:num]
	_, err = io.ReadFull(src, b)
	if err != nil {
		pool.Put(b)
		return err
	}

	message := bytes.NewReader(b)

	fm.Payload, err = ReadBytes(message)
	if err != nil {
		pool.Put(b)
		return err
	}

	pool.Put(b)
	return nil
}

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
	FramePool.Put(f.Payload)
}

var FramePool = NewPool(256, 1024, 8*1024)
