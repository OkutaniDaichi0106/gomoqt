package message

import (
	"io"
)

/*
 * Frame Message {
 *   Payload ([]byte),
 * }
 */

type FrameMessage struct {
	Payload []byte
}

func (fm FrameMessage) Len() int {
	return len(fm.Payload)
}

func (fm FrameMessage) Encode(w io.Writer) error {
	err := WriteVarint(w, uint64(len(fm.Payload)))
	if err != nil {
		return err
	}

	_, err = w.Write(fm.Payload)
	if err != nil {
		return err
	}

	return nil
}

func (fm *FrameMessage) Decode(src io.Reader) error {
	num, err := ReadVarint(src)
	if err != nil {
		return err
	}

	// If payload length is zero, reset the slice to zero length
	if num == 0 {
		fm.Payload = fm.Payload[:0]
		return nil
	}

	// Ensure the payload slice has enough capacity
	if cap(fm.Payload) < int(num) {
		fm.Payload = make([]byte, num)
	} else {
		fm.Payload = fm.Payload[:num]
	}

	_, err = io.ReadFull(src, fm.Payload)

	return err
}
