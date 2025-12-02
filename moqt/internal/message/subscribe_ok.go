package message

import (
	"io"
)

/*
 * SUBSCRIBE_OK Message {
 *   Group Frequency (varint),
 * }
 */
type SubscribeOkMessage struct {
}

func (som SubscribeOkMessage) Len() int {
	return 0
}

func (som SubscribeOkMessage) Encode(w io.Writer) error {
	msgLen := som.Len()
	b := pool.Get(msgLen + VarintLen(uint64(msgLen)))
	defer pool.Put(b)

	b, _ = WriteMessageLength(b, uint16(msgLen))

	_, err := w.Write(b)

	return err
}

func (som *SubscribeOkMessage) Decode(src io.Reader) error {
	num, err := ReadMessageLength(src)
	if err != nil {
		return err
	}

	b := pool.Get(int(num))[:num]
	defer pool.Put(b)
	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}
