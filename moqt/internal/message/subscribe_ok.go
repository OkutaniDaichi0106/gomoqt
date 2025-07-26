package message

import (
	"bytes"
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

	b = quicvarint.Append(b, uint64(msgLen))
	b = quicvarint.Append(b, uint64(som.GroupOrder))

	_, err := w.Write(b)

	pool.Put(b)

	return err
}

func (som *SubscribeOkMessage) Decode(src io.Reader) error {
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

	r := bytes.NewReader(b)

	num, err = ReadVarint(r)
	if err != nil {
		pool.Put(b)
		return err
	}
	som.GroupOrder = GroupOrder(num)

	pool.Put(b)

	return nil
}
