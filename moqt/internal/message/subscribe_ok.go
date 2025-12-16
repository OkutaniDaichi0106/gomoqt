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
	b := make([]byte, 0, msgLen+VarintLen(uint64(msgLen)))

	b, _ = WriteMessageLength(b, uint64(msgLen))

	_, err := w.Write(b)

	return err
}

func (som *SubscribeOkMessage) Decode(src io.Reader) error {
	num, err := ReadMessageLength(src)
	if err != nil {
		return err
	}

	b := make([]byte, num)
	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}
