package moqt

import (
	"errors"
	"fmt"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

var (
	// ErrInvalidScheme is returned when a URL scheme is not supported.
	// Only "https" (for WebTransport) and "moqt" (for QUIC) schemes are valid.
	ErrInvalidScheme = errors.New("moqt: invalid scheme")

	// ErrInvalidRange is returned when a subscribe range is invalid.
	ErrInvalidRange = errors.New("moqt: invalid range")

	// ErrClosedSession is returned when attempting to use a closed session.
	ErrClosedSession = errors.New("moqt: closed session")

	// ErrServerClosed is returned when the server has been closed.
	ErrServerClosed = errors.New("moqt: server closed")

	// ErrClientClosed is returned when the client has been closed.
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

var AnnounceErrorCodeTexts = map[AnnounceErrorCode]string{
	InternalAnnounceErrorCode:      "moqt: internal error",
	DuplicatedAnnounceErrorCode:    "moqt: duplicated broadcast path",
	InvalidAnnounceStatusErrorCode: "moqt: invalid announce status",
	UninterestedErrorCode:          "moqt: uninterested",
	BannedPrefixErrorCode:          "moqt: banned prefix",
	InvalidPrefixErrorCode:         "moqt: invalid prefix",
}

// AnnounceErrorCode represents error codes for track announcement operations.
// These codes are used when an announcement is rejected or fails.
type AnnounceErrorCode quic.StreamErrorCode

func (code AnnounceErrorCode) String() string {
	return AnnounceErrorCodeTexts[code]
}

// AnnounceError wraps a QUIC stream error with announcement-specific error codes.
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

var SubscribeErrorCodeTexts = map[SubscribeErrorCode]string{
	InternalSubscribeErrorCode:     "moqt: internal error",
	InvalidRangeErrorCode:          "moqt: invalid range",
	DuplicateSubscribeIDErrorCode:  "moqt: duplicated id",
	TrackNotFoundErrorCode:         "moqt: track does not exist",
	UnauthorizedSubscribeErrorCode: "moqt: unauthorized",
	SubscribeTimeoutErrorCode:      "moqt: timeout",
}

// SubscribeErrorCode represents error codes for subscription operations.
// These codes are used when a subscription request is rejected or fails.
type SubscribeErrorCode quic.StreamErrorCode

func (code SubscribeErrorCode) String() string {
	return SubscribeErrorCodeTexts[code]
}

// SubscribeError wraps a QUIC stream error with subscription-specific error codes.
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

	SetupFailedErrorCode SessionErrorCode = 0x13
)

var SessionErrorCodeTexts = map[SessionErrorCode]string{
	NoError:                          "moqt: no error",
	InternalSessionErrorCode:         "moqt: internal error",
	UnauthorizedSessionErrorCode:     "moqt: unauthorized",
	ProtocolViolationErrorCode:       "moqt: protocol violation",
	ParameterLengthMismatchErrorCode: "moqt: parameter length mismatch",
	TooManySubscribeErrorCode:        "moqt: too many subscribes",
	GoAwayTimeoutErrorCode:           "moqt: goaway timeout",
	UnsupportedVersionErrorCode:      "moqt: unsupported version",
	SetupFailedErrorCode:             "moqt: setup failed",
}

// SessionErrorCode represents error codes for MOQ session operations.
// These codes are used at the connection level for protocol errors.
type SessionErrorCode quic.ApplicationErrorCode

func (code SessionErrorCode) String() string {
	return SessionErrorCodeTexts[code]
}

// SessionError wraps a QUIC application error with session-specific error codes.
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

var GroupErrorCodeTexts = map[GroupErrorCode]string{
	InternalGroupErrorCode:      "moqt: internal error",
	OutOfRangeErrorCode:         "moqt: out of range",
	ExpiredGroupErrorCode:       "moqt: group expires",
	SubscribeCanceledErrorCode:  "moqt: subscribe canceled",
	PublishAbortedErrorCode:     "moqt: publish aborted",
	ClosedSessionGroupErrorCode: "moqt: session closed",
	InvalidSubscribeIDErrorCode: "moqt: invalid subscribe id",
}

type GroupErrorCode quic.StreamErrorCode

func (code GroupErrorCode) String() string {
	return GroupErrorCodeTexts[code]
}

type GroupError struct{ *quic.StreamError }

func (err GroupError) Error() string {
	return err.GroupErrorCode().String()
}

func (err GroupError) GroupErrorCode() GroupErrorCode {
	return GroupErrorCode(err.ErrorCode)
}

// /*
//  * Internal Error
//  */
// type InternalError struct {
// 	Reason string
// }

// func (err InternalError) Error() string {
// 	return fmt.Sprintf("moqt: internal error: %s", err.Reason)
// }

// func (InternalError) Is(err error) bool {
// 	_, ok := err.(InternalError)
// 	return ok
// }

// func (err InternalError) AnnounceErrorCode() AnnounceErrorCode {
// 	return InternalAnnounceErrorCode
// }

// func (err InternalError) SubscribeErrorCode() SubscribeErrorCode {
// 	return InternalSubscribeErrorCode
// }

// func (err InternalError) SessionErrorCode() SessionErrorCode {
// 	return InternalSessionErrorCode
// }

// func (err InternalError) GroupErrorCode() GroupErrorCode {
// 	return InternalGroupErrorCode
// }

// /*
//  * Unauthorized Error
//  */
// type UnauthorizedError struct{}

// func (UnauthorizedError) Error() string {
// 	return "moqt: unauthorized"
// }

// func (UnauthorizedError) Is(err error) bool {
// 	_, ok := err.(UnauthorizedError)
// 	return ok
// }

// func (err UnauthorizedError) SubscribeErrorCode() SubscribeErrorCode {
// 	return UnauthorizedSubscribeErrorCode
// }

// func (err UnauthorizedError) SessionErrorCode() SessionErrorCode {
// 	return UnauthorizedSessionErrorCode
// }
