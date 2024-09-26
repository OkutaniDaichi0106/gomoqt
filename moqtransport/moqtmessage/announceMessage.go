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
	 * Serialize as following formatt
	 *
	 * ANNOUNCE Payload {
	 *   Track Namespace (tuple),
	 *   Number of Parameters (),
	 *   Announce Parameters(..)
	 * }
	 */

	// TODO: Tune the size of the slice
	b := make([]byte, 0, 1<<8)

	// Append message ID
	b = quicvarint.Append(b, uint64(ANNOUNCE))

	// Append the Track Namespace
	b = a.TrackNamespace.Append(b)

	// Serialize the parameters and append it
	/*
	 * Announce Parameters {
	 *   [Authorization Info Parameter (stirng)],
	 *   [Optional Patameters(..)],
	 * }
	 */
	b = a.Parameters.append(b)

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
