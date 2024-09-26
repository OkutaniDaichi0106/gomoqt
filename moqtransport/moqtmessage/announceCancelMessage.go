package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceCancelCode uint64

const (
	ANNOUNCE_CANCEL_NO_ERROR AnnounceCancelCode = 0x0
)

type AnnounceCancelMessage struct {
	TrackNamespace TrackNamespace
	ErrorCode      AnnounceCancelCode
	Reason         string
}

func (ac AnnounceCancelMessage) Serialize() []byte {
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

	// Append the error code of the cancel
	b = quicvarint.Append(b, uint64(ac.ErrorCode))

	// Append the reason of the cancel
	b = quicvarint.Append(b, uint64(len(ac.Reason)))
	b = append(b, []byte(ac.Reason)...)

	// Append the supported versions
	b = ac.TrackNamespace.Append(b)

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

func (ac *AnnounceCancelMessage) DeserializeBody(r quicvarint.Reader) error {
	// Get Track Namespace
	if ac.TrackNamespace == nil {
		ac.TrackNamespace = make(TrackNamespace, 0, 1)
	}

	err := ac.TrackNamespace.Deserialize(r)

	return err
}
