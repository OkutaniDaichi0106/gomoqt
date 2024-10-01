package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceMessage struct {
	/*
	 * Track Namespace
	 */
	TrackNamespace TrackNamespace

	/*
	 * Announce Parameters
	 * Parameters should include track authorization information
	 */
	Parameters Parameters
}

func (a AnnounceMessage) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * ANNOUNCE Message {
	 *   Type (varint) = 0x06,
	 *   Length (varint),
	 *   Track Namespace (tuple),
	 *   Number of Parameters (),
	 *   Announce Parameters(..)
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Track Namespace
	p = a.TrackNamespace.Append(p)

	// Append the Parameters
	p = a.Parameters.append(p)

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the message type
	b = quicvarint.Append(b, uint64(ANNOUNCE))

	// Append the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (a *AnnounceMessage) DeserializeBody(r quicvarint.Reader) error {
	var tns TrackNamespace
	err := tns.Deserialize(r)
	if err != nil {
		return err
	}

	a.TrackNamespace = tns

	err = a.Parameters.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}
