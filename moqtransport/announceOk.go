package moqtransport

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceOkMessage struct {
	TrackNamespace TrackNamespace
}

func (ao AnnounceOkMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * ANNOUNCE_OK Payload {
	 *   Track Namespace ([]byte),
	 * }
	 */

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */

	// Append the type of the message
	b = quicvarint.Append(b, uint64(ANNOUNCE_OK))

	// Append the supported versions
	b = ao.TrackNamespace.append(b)

	return b
}

func (ao *AnnounceOkMessage) deserializeBody(r quicvarint.Reader) error {
	// Get Track Namespace
	var tns TrackNamespace
	err := tns.deserialize(r)
	if err != nil {
		return err
	}

	ao.TrackNamespace = tns

	return nil
}
