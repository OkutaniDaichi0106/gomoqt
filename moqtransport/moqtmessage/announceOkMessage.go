package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceOkMessage struct {
	TrackNamespace TrackNamespace
}

func (ao AnnounceOkMessage) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * ANNOUNCE_OK Message {
	 *   Type (varint) = 0x07,
	 *   Length (varint),
	 *   Track Namespace (tuple),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<8)

	// Append the supported versions
	p = ao.TrackNamespace.Append(p)

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the type of the message
	b = quicvarint.Append(b, uint64(ANNOUNCE_OK))

	// Append the length of the payload and the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (ao *AnnounceOkMessage) DeserializePayload(r quicvarint.Reader) error {
	// Get Track Namespace
	var tns TrackNamespace
	err := tns.Deserialize(r)
	if err != nil {
		return err
	}
	ao.TrackNamespace = tns

	return nil
}
