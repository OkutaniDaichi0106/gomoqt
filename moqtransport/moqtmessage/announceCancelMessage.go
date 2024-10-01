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
	 * Serialize the message in the following formatt
	 *
	 * ANNOUNCE_CANCEL Message {
	 *   Type (varint) = 0x0C,
	 *   Length (varint),
	 *   Track Namespace (tuple),
	 *   Error Code (varint),
	 *   Reason Phrase (string),
	 * }
	 */

	// TODO?: Check track namespace exists

	// TODO: Tune the length of the "b"
	p := make([]byte, 0, 1<<8)

	// Append the Track Namespace
	p = ac.TrackNamespace.Append(p)

	// Append the error code of the cancel
	p = quicvarint.Append(p, uint64(ac.ErrorCode))

	// Append the reason of the cancelation
	p = quicvarint.Append(p, uint64(len(ac.Reason)))
	p = append(p, []byte(ac.Reason)...)

	/*
	 * Serialize the whole data
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the type of the message
	b = quicvarint.Append(b, uint64(ANNOUNCE_CANCEL))

	// Append the length of the payload and the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (ac *AnnounceCancelMessage) DeserializePayload(r quicvarint.Reader) error {
	// Get the Track Namespace
	var tns TrackNamespace
	err := tns.Deserialize(r)
	if err != nil {
		return err
	}
	ac.TrackNamespace = tns

	// Get the Error Code
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}
	ac.ErrorCode = AnnounceCancelCode(num)

	// Get the Error Reason Phrase
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	ac.Reason = string(buf)

	return nil
}
