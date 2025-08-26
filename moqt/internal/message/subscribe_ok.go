package message

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * SUBSCRIBE_OK Message {
 *   Group Order (varint),
 * }
 */
type SubscribeOkMessage struct {
	GroupOrder GroupOrder
}

func (som SubscribeOkMessage) Len() int {
	return VarintLen(uint64(som.GroupOrder))
}

func (som SubscribeOkMessage) Encode(w io.Writer) error {
	msgLen := som.Len()
	b := pool.Get(msgLen)
	defer pool.Put(b)

	b = quicvarint.Append(b, uint64(msgLen))
	b = quicvarint.Append(b, uint64(som.GroupOrder))

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

	num, n, err := ReadVarint(b)
	if err != nil {
		return err
	}
	som.GroupOrder = GroupOrder(num)
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}
