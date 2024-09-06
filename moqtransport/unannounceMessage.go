package moqtransport

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type UnannounceMessage struct {
	TrackNamespace string
}

func (ua UnannounceMessage) serialize() []byte {
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
	// Append the supported versions
	b = quicvarint.Append(b, uint64(len(ua.TrackNamespace)))
	b = append(b, []byte(ua.TrackNamespace)...)

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

func (ua *UnannounceMessage) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64
	// Get length of the string of the track namespace
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get track namespace
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}

	ua.TrackNamespace = string(buf)

	// Just one track namespace supposed to be detected
	// Over one track namespace will not be detected

	return nil
}
