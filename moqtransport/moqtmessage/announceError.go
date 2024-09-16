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
	 * Serialize as following formatt
	 *
	 * ANNOUNCE_ERROR Message {
	 *   Track Namespace ([]byte),
	 *   Error Code (varint),
	 *   Reason Phrase ([]byte]),
	 * }
	 */

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<8) /* Byte slice storing whole data */

	// Append the type of the message
	b = quicvarint.Append(b, uint64(ANNOUNCE_ERROR))

	// Append Subscriber ID
	b = ae.TrackNamespace.Append(b)

	// Append Error Code
	b = quicvarint.Append(b, uint64(ae.Code))

	// Append Reason
	b = quicvarint.Append(b, uint64(len(ae.Reason)))
	b = append(b, []byte(ae.Reason)...)

	return b
}

func (ae *AnnounceErrorMessage) DeserializeBody(r quicvarint.Reader) error {
	var num uint64

	// Get Track Namespace
	if ae.TrackNamespace == nil {
		ae.TrackNamespace = make(TrackNamespace, 0, 1)
	}
	err := ae.TrackNamespace.Deserialize(r)
	if err != nil {
		return err
	}

	// Get Error Code
	num, err = quicvarint.Read(r)
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
