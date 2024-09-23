package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type UnannounceMessage struct {
	TrackNamespace
}

func (ua UnannounceMessage) Serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * UNANNOUNCE Payload {
	 *   Track Namespace ([]byte),
	 * }
	 */

	// TODO?: Chech track namespace exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(UNANNOUNCE))
	// Append the Track Namespace
	b = ua.TrackNamespace.Append(b)

	return b
}

// func (ua *UnannounceMessage) deserialize(r quicvarint.Reader) error {
// 	// Get Message ID and check it
// 	id, err := deserializeHeader(r)
// 	if err != nil {
// 		return err
// 	}
// 	if id != UNANNOUNCE {
// 		return errors.New("unexpected message")
// 	}

// 	return ua.deserializeBody(r)
// }

func (ua *UnannounceMessage) DeserializeBody(r quicvarint.Reader) error {
	var tns TrackNamespace
	err := tns.Deserialize(r)
	if err != nil {
		return err
	}

	ua.TrackNamespace = tns

	return nil
}
