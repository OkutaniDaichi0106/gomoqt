package moqtmessage

import "github.com/quic-go/quic-go/quicvarint"

type UnsubscribeNamespace struct {
	TrackNamespacePrefix TrackNamespacePrefix
}

func (usn UnsubscribeNamespace) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * UNSUBSCRIBE_NAMESPACE Message {
	 *   Type (varint) = 0x14,
	 *   Length (varint),
	 *   Track Namespace Prefix (tuple),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Track Namespace Prefix
	p = usn.TrackNamespacePrefix.Append(p)

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the message ID
	b = quicvarint.Append(b, uint64(UNSUBSCRIBE_NAMESPACE))

	// Append the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (usn *UnsubscribeNamespace) Deserialize(r quicvarint.Reader) error {
	if usn.TrackNamespacePrefix == nil {
		usn.TrackNamespacePrefix = make(TrackNamespacePrefix, 0, 1)
	}
	err := usn.TrackNamespacePrefix.Deserialize(r)

	return err
}
