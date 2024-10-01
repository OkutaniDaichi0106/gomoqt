package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceErrorCode int

const (
	ANNOUNCE_INTERNAL_ERROR   AnnounceErrorCode = 0x0 // Original
	DUPLICATE_TRACK_NAMESPACE AnnounceErrorCode = 0x1 // Original
)

/*
 * Subscribers sends ANNOUNCE_ERROR control message for tracks that failed authorization
 */
type AnnounceErrorMessage struct {
	TrackNamespace TrackNamespace
	Code           AnnounceErrorCode
	Reason         string
}

func (ae AnnounceErrorMessage) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * ANNOUNCE_ERROR Message {
	 *   Track Namespace (tuple),
	 *   Error Code (varint),
	 *   Reason Phrase ([]byte]),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Subscriber ID
	p = ae.TrackNamespace.Append(p)

	// Append the Error Code
	p = quicvarint.Append(p, uint64(ae.Code))

	// Append the Reason Phrase
	p = quicvarint.Append(p, uint64(len(ae.Reason)))
	p = append(p, []byte(ae.Reason)...)

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the message type
	b = quicvarint.Append(b, uint64(ANNOUNCE_ERROR))

	// Append the length of the payload and the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (ae *AnnounceErrorMessage) DeserializePayload(r quicvarint.Reader) error {
	// Get Track Namespace
	var tns TrackNamespace

	err := tns.Deserialize(r)
	if err != nil {
		return err
	}
	ae.TrackNamespace = tns

	// Get Error Code
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}
	ae.Code = AnnounceErrorCode(num)

	// Get Reason
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	ae.Reason = string(buf)

	return nil
}
