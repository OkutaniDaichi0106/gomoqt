package message

import (
	"io"

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

func (a AnnounceMessage) Encode(w io.Writer) error {
	/*
	 * Serialize the payload in the following format
	 *
	 * ANNOUNCE Message Payload {
	 *   Track Namespace (tuple),
	 *   Number of Parameters (),
	 *   Announce Parameters(..)
	 * }
	 */

	p := make([]byte, 0, 1<<6) // TODO: Tune the size

	// Append the Track Namespace
	p = appendTrackNamespace(p, a.TrackNamespace)

	// Append the Parameters
	p = appendParameters(p, a.Parameters)

	// Get serialized message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	_, err := w.Write(b)

	return err
}

func (am *AnnounceMessage) Decode(r io.Reader) error {
	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Track Namespace
	tns, err := readTrackNamespace(mr)
	if err != nil {
		return err
	}
	am.TrackNamespace = tns

	am.Parameters, err = readParameters(mr)
	if err != nil {
		return err
	}

	return nil
}
