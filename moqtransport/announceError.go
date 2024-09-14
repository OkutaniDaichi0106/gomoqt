package moqtransport

import "github.com/quic-go/quic-go/quicvarint"

/*
 * Error codes for annoncement failure
 *
 * The following error codes are defined in the official document
 * INTERNAL_ERROR
 * INVALID_RANGE
 * RETRY_TRACK_ALIAS
 */
const (
	ANNOUNCE_INTERNAL_ERROR   AnnounceErrorCode = 0x0 // Original
	DUPLICATE_TRACK_NAMESPACE AnnounceErrorCode = 0x1 // Original
)

var ANNOUNCE_ERROR_REASON = map[AnnounceErrorCode]string{
	ANNOUNCE_INTERNAL_ERROR:   "Internal Error",
	DUPLICATE_TRACK_NAMESPACE: "Duplicate Track Namespace",
}

type AnnounceErrorCode int

/*
 * Subscribers sends ANNOUNCE_ERROR control message for tracks that failed authorization
 */
type AnnounceError struct {
	TrackNamespace TrackNamespace
	Code           AnnounceErrorCode
	Reason         string
}

func (ae AnnounceError) serialize() []byte {
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
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */

	// Append the type of the message
	b = quicvarint.Append(b, uint64(ANNOUNCE_ERROR))

	// Append Subscriber ID
	b = ae.TrackNamespace.append(b)

	// Append Error Code
	b = quicvarint.Append(b, uint64(ae.Code))

	// Append Reason
	b = quicvarint.Append(b, uint64(len(ae.Reason)))
	b = append(b, []byte(ae.Reason)...)

	return b
}

func (ae *AnnounceError) deserializeBody(r quicvarint.Reader) error {
	var num uint64

	// Get Track Namespace
	if ae.TrackNamespace == nil {
		ae.TrackNamespace = make(TrackNamespace, 0, 1)
	}
	err := ae.TrackNamespace.deserialize(r)
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
