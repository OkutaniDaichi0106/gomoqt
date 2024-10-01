package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * Subscribe Error
 */
type SubscribeErrorCode uint

type SubscribeErrorMessage struct {
	/*
	 * A number to identify the subscribe session
	 */
	SubscribeID

	/*
	 * Error code
	 */
	Code SubscribeErrorCode

	/*
	 * Reason of the error
	 */
	Reason string

	/*
	 * An number indicates a track
	 * This is referenced instead of the Track Name and Track Namespace
	 */
	TrackAlias
}

// Error codes defined at official document
const (
	SUBSCRIBE_INTERNAL_ERROR SubscribeErrorCode = 0x0
	INVALID_RANGE            SubscribeErrorCode = 0x1
	RETRY_TRACK_ALIAS        SubscribeErrorCode = 0x2
	TRACK_DOES_NOT_EXIST     SubscribeErrorCode = 0x3
	SUBSCRIBE_UNAUTHORIZED   SubscribeErrorCode = 0x4
	SUBSCRIBE_TIMEOUT        SubscribeErrorCode = 0x5
)

func (se SubscribeErrorMessage) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * SUBSCRIBE_ERROR Message {
	 *   Type (varint) = 0x05,
	 *   Length (varint),
	 *   Subscribe ID (varint),
	 *   ErrorCode (varint),
	 *   Reason ([]byte]),
	 *   TrackAlias (varint),
	 * }
	 */

	p := make([]byte, 0, 1<<8)

	// Append the Subscriber ID
	p = quicvarint.Append(p, uint64(se.SubscribeID))

	// Append the Error Code
	p = quicvarint.Append(p, uint64(se.Code))

	// Append the Reason Phrase
	p = quicvarint.Append(p, uint64(len(se.Reason)))
	p = append(p, []byte(se.Reason)...)

	// Append the Track Alias
	p = quicvarint.Append(p, uint64(se.TrackAlias))

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the type of the message
	b = quicvarint.Append(b, uint64(SUBSCRIBE_ERROR))

	// Appen the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (se *SubscribeErrorMessage) DeserializePayload(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	se.SubscribeID = SubscribeID(num)

	// Get Error Code
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	se.Code = SubscribeErrorCode(num)

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
	se.Reason = string(buf)

	// Get Track Alias
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	se.TrackAlias = TrackAlias(num)

	return nil
}
