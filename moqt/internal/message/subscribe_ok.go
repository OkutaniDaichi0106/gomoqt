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

func (som SubscribeOkMessage) Encode(w io.Writer) error {
	p := getBytes()
	defer putBytes(p)

	p = AppendNumber(p, uint64(som.GroupOrder))

	_, err := w.Write(p)
	return err
}

func (som *SubscribeOkMessage) Decode(r io.Reader) error {
	num, _, err := ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}
	som.GroupOrder = GroupOrder(num)

	return nil
}
