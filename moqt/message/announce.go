package message

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

func (a AnnounceMessage) SerializePayload() []byte {
	/*
	 * Serialize the payload in the following formatt
	 *
	 * ANNOUNCE Message Payload {
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
	p = AppendTrackNamespace(p, a.TrackNamespace)

	// Append the Parameters
	p = a.Parameters.Append(p)

	return p
}

func (a *AnnounceMessage) DeserializePayload(r quicvarint.Reader) error {
	tns, err := ReadTrackNamespace(r)
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
