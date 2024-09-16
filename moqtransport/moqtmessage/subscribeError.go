package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * Subscribe Error
 */
type SubscribeErrorCode uint

type SubscribeError struct {
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
)

func (se SubscribeError) Serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * SUBSCRIBE_ERROR Message {
	 *   Subscribe ID (varint),
	 *   ErrorCode (varint),
	 *   Reason ([]byte]),
	 *   TrackAlias (varint),
	 * }
	 */

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */

	// Append the type of the message
	b = quicvarint.Append(b, uint64(SUBSCRIBE))

	// Append Subscriber ID
	b = quicvarint.Append(b, uint64(se.SubscribeID))

	// Append Error Code
	b = quicvarint.Append(b, uint64(se.Code))

	// Append Reason
	b = quicvarint.Append(b, uint64(len(se.Reason)))
	b = append(b, []byte(se.Reason)...)

	// Append Track Alias
	b = quicvarint.Append(b, uint64(se.TrackAlias))

	return b
}

func (se *SubscribeError) DeserializeBody(r quicvarint.Reader) error {
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
