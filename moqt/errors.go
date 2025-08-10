package moqt

import (
	"errors"
	"fmt"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

var (
	ErrInvalidScheme = errors.New("moqt: invalid scheme")

	ErrInvalidRange = errors.New("moqt: invalid range")

	ErrClosedSession = errors.New("moqt: closed session")

	ErrServerClosed = errors.New("moqt: server closed")

	ErrClientClosed = errors.New("moqt: client closed")
)

/*
 * Announce Errors
 */
const (
	InternalAnnounceErrorCode AnnounceErrorCode = 0x0

	// Subscriber
	DuplicatedAnnounceErrorCode    AnnounceErrorCode = 0x1
	InvalidAnnounceStatusErrorCode AnnounceErrorCode = 0x2 // TODO: Is this necessary?
	UninterestedErrorCode          AnnounceErrorCode = 0x3

	// Publisher
	BannedPrefixErrorCode  AnnounceErrorCode = 0x4 // TODO: Is this necessary?
	InvalidPrefixErrorCode AnnounceErrorCode = 0x5 // TODO: Is this necessary?
)

type AnnounceErrorCode quic.StreamErrorCode

func (code AnnounceErrorCode) String() string {
	switch code {
	case InternalAnnounceErrorCode:
		return "moqt: internal error"
	case DuplicatedAnnounceErrorCode:
		return "moqt: duplicated broadcast path"
	case UninterestedErrorCode:
		return "moqt: uninterested"
	default:
		return "moqt: unknown announce error"
	}
}

type AnnounceError struct{ *quic.StreamError }

func (err AnnounceError) Error() string {
	return err.AnnounceErrorCode().String()
}

func (err AnnounceError) AnnounceErrorCode() AnnounceErrorCode {
	return AnnounceErrorCode(err.ErrorCode)
}

/*
 * Subscribe Errors
 */

const (
	InternalSubscribeErrorCode SubscribeErrorCode = 0x00

	// Error code used internally, basically not used in the application.
	// These error codes are used before subscribe negotiation is completed,

	//
	InvalidRangeErrorCode SubscribeErrorCode = 0x01
	//
	DuplicateSubscribeIDErrorCode SubscribeErrorCode = 0x02
	//
	TrackNotFoundErrorCode SubscribeErrorCode = 0x03
	//
	UnauthorizedSubscribeErrorCode SubscribeErrorCode = 0x04 // TODO: Is this necessary?
	// Subscriber
	SubscribeTimeoutErrorCode SubscribeErrorCode = 0x05
	// ClosedTrackErrorCode           SubscribeErrorCode = 0x07 // TODO: Is this necessary?

	// Error code used by the application.
	// These error codes are used after subscribe negotiation is completed.
)

type SubscribeErrorCode quic.StreamErrorCode

func (code SubscribeErrorCode) String() string {
	switch code {
	case InternalSubscribeErrorCode:
		return "moqt: internal error"
	case InvalidRangeErrorCode:
		return "moqt: invalid range"
	case DuplicateSubscribeIDErrorCode:
		return "moqt: duplicated id"
	case TrackNotFoundErrorCode:
		return "moqt: track does not exist"
	case UnauthorizedSubscribeErrorCode:
		return "moqt: unauthorized"
	case SubscribeTimeoutErrorCode:
		return "moqt: timeout"
	default:
		return "moqt: unknown subscribe error"
	}
}

type SubscribeError struct{ *quic.StreamError }

func (err SubscribeError) Error() string {
	return err.SubscribeErrorCode().String()
}

func (err SubscribeError) SubscribeErrorCode() SubscribeErrorCode {
	return SubscribeErrorCode(err.ErrorCode)
}

/*
 * Session Error
 */
const (
	NoError SessionErrorCode = 0x0

	InternalSessionErrorCode         SessionErrorCode = 0x1
	UnauthorizedSessionErrorCode     SessionErrorCode = 0x2
	ProtocolViolationErrorCode       SessionErrorCode = 0x3
	ParameterLengthMismatchErrorCode SessionErrorCode = 0x5
	TooManySubscribeErrorCode        SessionErrorCode = 0x6
	GoAwayTimeoutErrorCode           SessionErrorCode = 0x10
	UnsupportedVersionErrorCode      SessionErrorCode = 0x12
)

type SessionErrorCode quic.ApplicationErrorCode

func (code SessionErrorCode) String() string {
	switch code {
	case NoError:
		return "moqt: no error"
	case InternalSessionErrorCode:
		return "moqt: internal error"
	case UnauthorizedSessionErrorCode:
		return "moqt: unauthorized"
	case ProtocolViolationErrorCode:
		return "moqt: protocol violation"
	case ParameterLengthMismatchErrorCode:
		return "moqt: parameter length mismatch"
	case TooManySubscribeErrorCode:
		return "moqt: too many subscribes"
	case GoAwayTimeoutErrorCode:
		return "moqt: goaway timeout"
	case UnsupportedVersionErrorCode:
		return "moqt: unsupported version"
	default:
		return "moqt: unknown session error"
	}
}

type SessionError struct{ *quic.ApplicationError }

func (err SessionError) Error() string {
	var role string
	if err.Remote {
		role = "remote"
	} else {
		role = "local"
	}
	return fmt.Sprintf("%s (%s)", err.SessionErrorCode().String(), role)
}

func (err SessionError) SessionErrorCode() SessionErrorCode {
	return SessionErrorCode(err.ErrorCode)
}

/*
 * Group Error
 */
const (
	InternalGroupErrorCode GroupErrorCode = 0x00

	OutOfRangeErrorCode         GroupErrorCode = 0x02
	ExpiredGroupErrorCode       GroupErrorCode = 0x03
	SubscribeCanceledErrorCode  GroupErrorCode = 0x04 // TODO: Is this necessary?
	PublishAbortedErrorCode     GroupErrorCode = 0x05
	ClosedSessionGroupErrorCode GroupErrorCode = 0x06
	InvalidSubscribeIDErrorCode GroupErrorCode = 0x07 // TODO: Is this necessary?
)

type GroupErrorCode quic.StreamErrorCode

func (code GroupErrorCode) String() string {
	switch code {
	case InternalGroupErrorCode:
		return "moqt: internal error"
	case OutOfRangeErrorCode:
		return "moqt: out of range"
	case ExpiredGroupErrorCode:
		return "moqt: group expires"
	case SubscribeCanceledErrorCode:
		return "moqt: subscribe canceled"
	case PublishAbortedErrorCode:
		return "moqt: publish aborted"
	case ClosedSessionGroupErrorCode:
		return "moqt: session closed"
	case InvalidSubscribeIDErrorCode:
		return "moqt: invalid subscribe id"
	default:
		return "moqt: unknown group error"
	}
}

type GroupError struct{ *quic.StreamError }

func (err GroupError) Error() string {
	return err.GroupErrorCode().String()
}

func (err GroupError) GroupErrorCode() GroupErrorCode {
	return GroupErrorCode(err.ErrorCode)
}

/*
 * Internal Error
 */
type InternalError struct {
	Reason string
}

func (err InternalError) Error() string {
	return fmt.Sprintf("moqt: internal error: %s", err.Reason)
}

func (InternalError) Is(err error) bool {
	_, ok := err.(InternalError)
	return ok
}

func (err InternalError) AnnounceErrorCode() AnnounceErrorCode {
	return InternalAnnounceErrorCode
}

func (err InternalError) SubscribeErrorCode() SubscribeErrorCode {
	return InternalSubscribeErrorCode
}

func (err InternalError) SessionErrorCode() SessionErrorCode {
	return InternalSessionErrorCode
}

func (err InternalError) GroupErrorCode() GroupErrorCode {
	return InternalGroupErrorCode
}

/*
 * Unauthorized Error
 */
type UnauthorizedError struct{}

func (UnauthorizedError) Error() string {
	return "moqt: unauthorized"
}

func (UnauthorizedError) Is(err error) bool {
	_, ok := err.(UnauthorizedError)
	return ok
}

func (err UnauthorizedError) SubscribeErrorCode() SubscribeErrorCode {
	return UnauthorizedSubscribeErrorCode
}

func (err UnauthorizedError) SessionErrorCode() SessionErrorCode {
	return UnauthorizedSessionErrorCode
}
