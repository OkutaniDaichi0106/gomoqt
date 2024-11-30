package message

import (
	"io"
	"log"

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
	p := make([]byte, 0, 1<<8) // TODO: Tune the size

	// Append the Track Namespace Prefix
	p = AppendTrackNamespacePrefix(p, aim.TrackPrefix)

	// Append the Parameters
	p = appendParameters(p, aim.Parameters)

	log.Print("ANNOUNCE_INTEREST payload", len(p)) // TODO: delete

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

func ReadAnnounceInterestMessage(r Reader) (AnnounceInterestMessage, error) {
	var aim AnnounceInterestMessage

	// Get a Track Namespace Prefix
	tnsp, err := ReadTrackNamespacePrefix(r)
	if err != nil {
		return AnnounceInterestMessage{}, err
	}
	aim.TrackPrefix = tnsp

	// Get Parameters
	aim.Parameters, err = readParameters(r)
	if err != nil {
		return AnnounceInterestMessage{}, err
	}

	return aim, nil
}
