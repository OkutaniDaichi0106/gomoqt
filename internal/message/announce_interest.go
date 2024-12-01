package message

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceInterestMessage struct {
	TrackPrefix TrackPrefix
	Parameters  Parameters
}

func (aim AnnounceInterestMessage) Encode(w io.Writer) error {
	/*
	 * Serialize the message in the following formatt
	 *
	 * ANNOUNCE_INTEREST Message Payload {
	 *   Track Namespace prefix (tuple),
	 *   Subscribe Parameters (Parameters),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<6) // TODO: Tune the size

	// Append the Track Namespace Prefix
	p = appendTrackNamespacePrefix(p, aim.TrackPrefix)

	// Append the Parameters
	p = appendParameters(p, aim.Parameters)

	/*
	 * Get serialized message
	 */
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)

	return err
}

func (aim *AnnounceInterestMessage) Decode(r io.Reader) error {
	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Track Namespace Prefix
	tnsp, err := readTrackNamespacePrefix(mr)
	if err != nil {
		return err
	}
	aim.TrackPrefix = tnsp

	// Get Parameters
	aim.Parameters, err = readParameters(mr)
	if err != nil {
		return err
	}

	return nil
}
