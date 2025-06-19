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
	return numberLen(uint64(som.GroupOrder))
}

func (som SubscribeOkMessage) Encode(w io.Writer) (int, error) {
	p := getBytes()
	defer putBytes(p)

	p = AppendNumber(p, uint64(som.Len()))
	p = AppendNumber(p, uint64(som.GroupOrder))

	return w.Write(p)
}

func (som *SubscribeOkMessage) Decode(r io.Reader) (int, error) {
	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := ReadNumber(mr)
	if err != nil {
		return n, err
	}
	som.GroupOrder = GroupOrder(num)

	return n, nil
}
