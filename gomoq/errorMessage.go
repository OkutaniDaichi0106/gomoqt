package gomoq

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type TerminateErrorCode int
type TerminateError struct {
}

/*
 * Error codes and status codes for termination of the session
 *
 * The following error codes and status codes are defined in the official document
 * NO_ERROR
 * INTERNAL_ERROR
 * UNAUTHORIZED
 * PROTOCOL_VIOLATION
 * DUPLICATE_TRACK_ALIAS
 * PARAMETER_LENGTH_MISMATCH
 * GOAWAY_TIMEOUT
 */
func (TerminateError) NO_ERROR() TerminateErrorCode {
	return 0x0
}

func (TerminateError) INTERNAL_ERROR() TerminateErrorCode {
	return 0x1
}

func (TerminateError) UNAUTHORIZED() TerminateErrorCode {
	return 0x2
}

func (TerminateError) PROTOCOL_VIOLATION() TerminateErrorCode {
	return 0x3
}

func (TerminateError) DUPLICATE_TRACK_ALIAS() TerminateErrorCode {
	return 0x4
}

func (TerminateError) PARAMETER_LENGTH_MISMATCH() TerminateErrorCode {
	return 0x5
}

func (TerminateError) GOAWAY_TIMEOUT() TerminateErrorCode {
	return 0x6
}

/*
 * Subscribe Error
 */
type AnnounceErrorCode int

/*
 * Subscribers sends ANNOUNCE_ERROR control message for tracks that failed authorization
 */
type AnnounceError struct {
	TrackNamespace string
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
	b = quicvarint.Append(b, uint64(len(ae.TrackNamespace)))
	b = append(b, []byte(ae.TrackNamespace)...)

	// Append Error Code
	b = quicvarint.Append(b, uint64(ae.Code))

	// Append Reason
	b = quicvarint.Append(b, uint64(len(ae.Reason)))
	b = append(b, []byte(ae.Reason)...)

	return b
}

func (ae *AnnounceError) deserialize(r quicvarint.Reader) error {
	// Get Message ID and check it
	id, err := deserializeHeader(r)
	if err != nil {
		return err
	}
	if id != ANNOUNCE_ERROR { //TODO: this would means protocol violation
		return errors.New("unexpected message")
	}

	return ae.deserializeBody(r)
}

func (ae *AnnounceError) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Track Namespace
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	ae.TrackNamespace = string(buf)

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
	buf = make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	ae.Reason = string(buf)

	return nil
}

/*
 * Error codes for annoncement failure
 *
 * The following error codes are defined in the official document
 * INTERNAL_ERROR
 * INVALID_RANGE
 * RETRY_TRACK_ALIAS
 */
func (AnnounceError) INTERNAL_ERROR() SubscribeErrorCode {
	return SubscribeErrorCode(0x0)
}

func (AnnounceError) INVALID_RANGE() SubscribeErrorCode {
	return SubscribeErrorCode(0x1)
}

func (AnnounceError) RETRY_TRACK_ALIAS() SubscribeErrorCode {
	return SubscribeErrorCode(0x2)
}

/*
 * Subscribe Error
 */
type SubscribeErrorCode int

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
func (SubscribeError) INTERNAL_ERROR() SubscribeErrorCode {
	return 0x0
}

func (SubscribeError) INVALID_RANGE() SubscribeErrorCode {
	return 0x1
}

func (SubscribeError) RETRY_TRACK_ALIAS() SubscribeErrorCode {
	return 0x2
}

func (se SubscribeError) serialize() []byte {
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

func (se *SubscribeError) deserialize(r quicvarint.Reader) error {
	// Get Message ID and check it
	id, err := deserializeHeader(r)
	if err != nil {
		return err
	}
	if id != SUBSCRIBE_ERROR {
		return ErrUnexpectedMessage
	}

	return nil
}

func (se *SubscribeError) deserializeBody(r quicvarint.Reader) error {
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
