package moqterror

import (
	"go-moq/moqtransport/moqtmessage"

	"github.com/quic-go/quic-go/quicvarint"
)

var (
	ErrAnnounceFailed = AnnounceInternalError{}
)

/*
 * Error codes for annoncement failure
 *
 * The following error codes are defined in the official document
 * INTERNAL_ERROR
 * INVALID_RANGE
 * RETRY_TRACK_ALIAS
 */
const (
	ANNOUNCE_INTERNAL_ERROR            AnnounceErrorCode = 0x0 // Original
	ANNOUNCE_DUPLICATE_TRACK_NAMESPACE AnnounceErrorCode = 0x1 // Original
)

type AnnounceError interface {
	error
	Code() AnnounceErrorCode
}

type AnnounceInternalError struct {
}

func (AnnounceInternalError) Error() string {
	return "internal error"
}

func (AnnounceInternalError) Code() AnnounceErrorCode {
	return ANNOUNCE_INTERNAL_ERROR
}

type AnnounceDuplicateTrackNamespace struct {
}

func (AnnounceDuplicateTrackNamespace) Error() string {
	return "duplicate track namespace"
}

func (AnnounceDuplicateTrackNamespace) Code() AnnounceErrorCode {
	return ANNOUNCE_DUPLICATE_TRACK_NAMESPACE
}

type AnnounceErrorCode int

/*
 * Subscribers sends ANNOUNCE_ERROR control message for tracks that failed authorization
 */
type AnnounceErrorMessage struct {
	TrackNamespace moqtmessage.TrackNamespace
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
	b = quicvarint.Append(b, uint64(moqtmessage.ANNOUNCE_ERROR))

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
		ae.TrackNamespace = make(moqtmessage.TrackNamespace, 0, 1)
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
