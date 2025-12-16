package message

import (
	"io"
)

type SessionUpdateMessage struct {
	/*
	 * Versions selected by the server
	 */
	Bitrate uint64
}

func (sum SessionUpdateMessage) Len() int {
	return VarintLen(sum.Bitrate)
}

func (sum SessionUpdateMessage) Encode(w io.Writer) error {
	msgLen := sum.Len()
	b := make([]byte, 0, msgLen+VarintLen(uint64(msgLen)))

	b, _ = WriteMessageLength(b, uint64(msgLen))
	b, _ = WriteVarint(b, sum.Bitrate)

	_, err := w.Write(b)
	return err
}

func (sum *SessionUpdateMessage) Decode(src io.Reader) error {
	size, err := ReadMessageLength(src)
	if err != nil {
		return err
	}

	b := make([]byte, size)

	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	num, n, err := ReadVarint(b)
	if err != nil {
		return err
	}
	sum.Bitrate = num
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}
