package moqtmessage

import "github.com/quic-go/quic-go/quicvarint"

type SubscribeNamespaceMessage struct {
	TrackNamespacePrefix
	Parameters Parameters
}

func (sn SubscribeNamespaceMessage) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * SUBSCRIBE_NAMESPACE Message {
	 *   Type (varint) = 0x11,
	 *   Length (varint),
	 *   Track Namespace prefix (tuple),
	 *   Number of Parameters (varint),
	 *   Subscribe Parameters (..),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Track Namespace Prefix
	p = sn.TrackNamespacePrefix.Append(p)

	// Append the Parameters
	p = sn.Parameters.append(p)

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the message type
	b = quicvarint.Append(b, uint64(SUBSCRIBE_NAMESPACE))

	// Append the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (sn *SubscribeNamespaceMessage) DeserializePayload(r quicvarint.Reader) error {
	// Get Track Namespace Prefix
	var tnsp TrackNamespacePrefix
	err := tnsp.Deserialize(r)
	if err != nil {
		return err
	}
	sn.TrackNamespacePrefix = tnsp

	// Get Parameters
	err = sn.Parameters.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}
