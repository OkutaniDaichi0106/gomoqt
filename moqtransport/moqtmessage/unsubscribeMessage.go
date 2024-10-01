package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type UnsubscribeMessage struct {
	/*
	 * A number to identify the subscribe session
	 */
	SubscribeID
}

func (us UnsubscribeMessage) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * UNSUBSCRIBE Message {
	 *   Type (varint) = 0x0A,
	 *   Length (varint),
	 *   Subscribe ID (varint),
	 * }
	 */

	p := make([]byte, 0, 1<<8)
	// Append the type of the message

	// Append the Subscirbe ID
	p = quicvarint.Append(p, uint64(us.SubscribeID))

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the message type
	b = quicvarint.Append(b, uint64(UNSUBSCRIBE))

	// Append the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (us *UnsubscribeMessage) DeserializePayload(r quicvarint.Reader) error {
	// Get Subscribe ID
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}
	us.SubscribeID = SubscribeID(num)

	return nil
}
