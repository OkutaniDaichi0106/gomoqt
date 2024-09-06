package moqtransport

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceCancelMessage struct {
	TrackNamespace string
}

func (ac AnnounceCancelMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * ANNOUNCE_CANCEL Payload {
	 *   Track Namespace ([]byte),
	 * }
	 */

	// TODO?: Check track namespace exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(ANNOUNCE_CANCEL))
	// Append the supported versions
	b = quicvarint.Append(b, uint64(len(ac.TrackNamespace)))
	b = append(b, []byte(ac.TrackNamespace)...)

	return b
}

// func (ac *AnnounceCancelMessage) deserialize(r quicvarint.Reader) error {
// 	// Get Message ID and check it
// 	id, err := deserializeHeader(r)
// 	if err != nil {
// 		return err
// 	}
// 	if id != ANNOUNCE_CANCEL {
// 		return errors.New("unexpected message")
// 	}

// 	return ac.deserializeBody(r)
// }

func (ac *AnnounceCancelMessage) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get length of the string of the namespace
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	buf := make([]byte, num)
	// Get track namespace
	_, err = r.Read(buf)
	if err != nil {
		return err
	}

	ac.TrackNamespace = string(buf)

	// Just one track namespace supposed to be detected
	// Over one track namespace will not be detected

	return nil
}
