package message

import (
	"io"
)

/*
 * Frame Message {
 *   Payload ([]byte),
 * }
 */
type FrameMessage []byte

func (fm FrameMessage) Len() int {
	return len(fm)
}

func (fm FrameMessage) Encode(w io.Writer) error {
	err := writeVarint(w, uint64(len(fm)))
	if err != nil {
		return err
	}

	_, err = w.Write(fm)
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
		*fm = (*fm)[:0]
		return nil
	}

	// Ensure the payload slice has enough capacity
	if cap(*fm) < int(num) {
		*fm = make([]byte, num)
	} else {
		*fm = (*fm)[:num]
	}

	_, err = io.ReadFull(src, *fm)

	return err
}
