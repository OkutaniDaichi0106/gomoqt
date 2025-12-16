package message

import (
	"io"
)

/*
 * SUBSCRIBE_UPDATE Message {
 *   Track Priority (varint),
 * }
 */
type SubscribeUpdateMessage struct {
	TrackPriority uint8
}

func (su SubscribeUpdateMessage) Len() int {
	var l int

	l += VarintLen(uint64(su.TrackPriority))

	return l
}

func (su SubscribeUpdateMessage) Encode(w io.Writer) error {
	msgLen := su.Len()
	p := make([]byte, 0, msgLen+VarintLen(uint64(msgLen)))

	p, _ = WriteMessageLength(p, uint64(msgLen))
	p, _ = WriteVarint(p, uint64(su.TrackPriority))

	_, err := w.Write(p)

	return err
}

func (sum *SubscribeUpdateMessage) Decode(src io.Reader) error {
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
	sum.TrackPriority = uint8(num)
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}
