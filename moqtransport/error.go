package moqtransport

import "github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"

/*
 * Announce Error
 */
type AnnounceErrorCode uint32

const (
	ANNOUNCE_INTERNAL_ERROR   AnnounceErrorCode = 0x0
	DUPLICATE_TRACK_NAMESPACE AnnounceErrorCode = 0x1
)

var (
	ErrDuplicatedTrackNamespace = DefaultAnnounceError{
		reason: "duplicate track namespace",
		code:   DUPLICATE_TRACK_NAMESPACE,
	}
)

type AnnounceError interface {
	error
	AnnounceErrorCode() AnnounceErrorCode
}

type DefaultAnnounceError struct {
	reason string
	code   AnnounceErrorCode
}

func (err DefaultAnnounceError) Error() string {
	return err.reason
}

func (err DefaultAnnounceError) Code() AnnounceErrorCode {
	return err.code
}

/*
 * Internal Error
 */
var _ AnnounceError = (*InternalError)(nil)
var _ SubscribeError = (*InternalError)(nil)
var _ SubscribeDoneError = (*InternalError)(nil)
var _ TerminateError = (*InternalError)(nil)

type InternalError struct{}

func (InternalError) Error() string {
	return "internal error"
}

func (InternalError) AnnounceErrorCode() AnnounceErrorCode {
	return ANNOUNCE_INTERNAL_ERROR
}

func (InternalError) SubscribeErrorCode() SubscribeErrorCode {
	return SUBSCRIBE_INTERNAL_ERROR
}

func (InternalError) SubscribeDoneErrorCode() SubscribeDoneStatusCode {
	return SUBSCRIBE_DONE_INTERNAL_ERROR
}

func (InternalError) TerminateErrorCode() TerminateErrorCode {
	return TERMINATE_INTERNAL_ERROR
}

var ErrInternalError = InternalError{}

/*
 * Unauthorized Error
 */
var _ SubscribeError = (*UnauthorizedError)(nil)
var _ SubscribeDoneError = (*UnauthorizedError)(nil)
var _ TerminateError = (*UnauthorizedError)(nil)

type UnauthorizedError struct{}

func (UnauthorizedError) Error() string {
	return "internal error"
}

func (UnauthorizedError) SubscribeErrorCode() SubscribeErrorCode {
	return SUBSCRIBE_UNAUTHORIZED
}

func (UnauthorizedError) SubscribeDoneErrorCode() SubscribeDoneStatusCode {
	return SUBSCRIBE_DONE_UNAUTHORIZED
}

func (UnauthorizedError) TerminateErrorCode() TerminateErrorCode {
	return TERMINATE_UNAUTHORIZED
}

var ErrUnauthorizedError = InternalError{}

/*
 * Subscribe Error
 */
type SubscribeErrorCode uint32

const (
	SUBSCRIBE_INTERNAL_ERROR       SubscribeErrorCode = 0x00
	SUBSCRIBE_INVALID_RANGE        SubscribeErrorCode = 0x01
	SUBSCRIBE_RETRY_TRACK_ALIAS    SubscribeErrorCode = 0x02
	SUBSCRIBE_TRACK_DOES_NOT_EXIST SubscribeErrorCode = 0x03
	SUBSCRIBE_UNAUTHORIZED         SubscribeErrorCode = 0x04
	SUBSCRIBE_TIMEOUT              SubscribeErrorCode = 0x05
)

/*
 * Subscribe Error
 */
var (
	ErrDefaultInvalidRange = DefaultSubscribeError{
		code:   SUBSCRIBE_INVALID_RANGE,
		reason: "invalid range",
	}

	ErrTrackDoesNotExist = DefaultSubscribeError{
		code:   SUBSCRIBE_TRACK_DOES_NOT_EXIST,
		reason: "track does not exist",
	}

	ErrSubscribeTimeout = DefaultSubscribeError{
		code:   SUBSCRIBE_TIMEOUT,
		reason: "time out",
	}
)

type SubscribeError interface {
	error
	SubscribeErrorCode() SubscribeErrorCode
}

type DefaultSubscribeError struct {
	code   SubscribeErrorCode
	reason string
}

func (err DefaultSubscribeError) Error() string {
	return err.reason
}

func (err DefaultSubscribeError) SubscribeErrorCode() SubscribeErrorCode {
	return err.code
}

type RetryTrackAliasError struct {
	reason     string
	trackAlias moqtmessage.TrackAlias
}

func (err RetryTrackAliasError) Error() string {
	return err.reason
}

func (err RetryTrackAliasError) SubscribeErrorCode() SubscribeErrorCode {
	return SUBSCRIBE_RETRY_TRACK_ALIAS
}

func (err RetryTrackAliasError) TrackAlias() moqtmessage.TrackAlias {
	return err.trackAlias
}

/*
 *
 */
type SubscribeDoneStatusCode uint32

const (
	SUBSCRIBE_DONE_UNSUBSCRIBED       SubscribeDoneStatusCode = 0x0
	SUBSCRIBE_DONE_INTERNAL_ERROR     SubscribeDoneStatusCode = 0x1
	SUBSCRIBE_DONE_UNAUTHORIZED       SubscribeDoneStatusCode = 0x2
	SUBSCRIBE_DONE_TRACK_ENDED        SubscribeDoneStatusCode = 0x3
	SUBSCRIBE_DONE_SUBSCRIPTION_ENDED SubscribeDoneStatusCode = 0x4
	SUBSCRIBE_DONE_GOING_AWAY         SubscribeDoneStatusCode = 0x5
	SUBSCRIBE_DONE_EXPIRED            SubscribeDoneStatusCode = 0x6
)

type SubscribeDoneError interface {
	error
	SubscribeDoneErrorCode() SubscribeDoneStatusCode
}

var (
	ErrSubscribeExpired = DefaultSubscribeDoneError{
		code:   SUBSCRIBE_DONE_EXPIRED,
		reason: "expired",
	}
)

/***/
var _ SubscribeDoneError = (*DefaultSubscribeDoneError)(nil)

type DefaultSubscribeDoneError struct {
	code   SubscribeDoneStatusCode
	reason string
}

func (err DefaultSubscribeDoneError) Error() string {
	return err.reason
}

func (err DefaultSubscribeDoneError) Reason() string {
	return err.reason
}

func (err DefaultSubscribeDoneError) SubscribeDoneErrorCode() SubscribeDoneStatusCode {
	return err.code
}

/***/
type SubscribeDoneStatus interface {
	Reason() string
	Code() SubscribeDoneStatusCode
}

var _ SubscribeDoneStatus = (*DefaultSubscribeDoneStatus)(nil)

var (
	StatusUnsubscribed = DefaultSubscribeDoneStatus{
		code:   SUBSCRIBE_DONE_UNSUBSCRIBED,
		reason: "unsubscribed",
	}
	StatusEndedTrack = DefaultSubscribeDoneStatus{
		code:   SUBSCRIBE_DONE_TRACK_ENDED,
		reason: "track ended",
	}
	StatusEndedSubscription = DefaultSubscribeDoneStatus{
		code:   SUBSCRIBE_DONE_SUBSCRIPTION_ENDED,
		reason: "subsription ended",
	}
	StatusGoingAway = DefaultSubscribeDoneStatus{
		code:   SUBSCRIBE_DONE_GOING_AWAY,
		reason: "going away",
	}
)

type DefaultSubscribeDoneStatus struct {
	code   SubscribeDoneStatusCode
	reason string
}

func (status DefaultSubscribeDoneStatus) Reason() string {
	return status.reason
}

func (status DefaultSubscribeDoneStatus) Code() SubscribeDoneStatusCode {
	return status.code
}

/***/

// type AnnounceCancelError interface {
// 	AnnounceCancelErrorCode() AnnounceCancelCode
// 	Reason() string
// }

// var _ AnnounceCancelError = (*DefaultAnnounceCancelError)(nil)

// type DefaultAnnounceCancelError struct {
// 	code   moqtmessage.AnnounceCancelCode
// 	reason string
// }

// func (cancel DefaultAnnounceCancelError) Code() moqtmessage.AnnounceCancelCode {
// 	return cancel.code
// }

// func (cancel DefaultAnnounceCancelError) Reason() string {
// 	return cancel.reason
// }

// type SubscribeNamespaceError interface {
// 	error
// 	Code() uint64
// }

// type DefaultSubscribeNamespaceError struct {
// 	code   moqtmessage.SubscribeNamespaceErrorCode
// 	reason string
// }

// func (err DefaultSubscribeNamespaceError) Error() string {
// 	return err.reason
// }

// func (err DefaultSubscribeNamespaceError) Code() moqtmessage.SubscribeNamespaceErrorCode {
// 	return err.code
// }

/*
 *
 */
type TerminateErrorCode int

var (
	NoTerminateErr = DefaultTerminateError{
		code:   TERMINATE_NO_ERROR,
		reason: "no error",
	}

	ErrProtocolViolation = DefaultTerminateError{
		code:   TERMINATE_PROTOCOL_VIOLATION,
		reason: "protocol violation",
	}
	ErrDuplicatedTrackAlias = DefaultTerminateError{
		code:   TERMINATE_DUPLICATE_TRACK_ALIAS,
		reason: "duplicate track alias",
	}
	ErrParameterLengthMismatch = DefaultTerminateError{
		code:   TERMINATE_PARAMETER_LENGTH_MISMATCH,
		reason: "parameter length mismatch",
	}
	ErrTooManySubscribes = DefaultTerminateError{
		code:   TERMINATE_TOO_MANY_SUBSCRIBES,
		reason: "too many subscribes",
	}
	ErrGoAwayTimeout = DefaultTerminateError{
		code:   TERMINATE_GOAWAY_TIMEOUT,
		reason: "goaway timeout",
	}
)

/*
 *
 */
const (
	TERMINATE_NO_ERROR                  TerminateErrorCode = 0x0
	TERMINATE_INTERNAL_ERROR            TerminateErrorCode = 0x1
	TERMINATE_UNAUTHORIZED              TerminateErrorCode = 0x2
	TERMINATE_PROTOCOL_VIOLATION        TerminateErrorCode = 0x3
	TERMINATE_DUPLICATE_TRACK_ALIAS     TerminateErrorCode = 0x4
	TERMINATE_PARAMETER_LENGTH_MISMATCH TerminateErrorCode = 0x5
	TERMINATE_TOO_MANY_SUBSCRIBES       TerminateErrorCode = 0x6
	TERMINATE_GOAWAY_TIMEOUT            TerminateErrorCode = 0x10
)

type TerminateError interface {
	error
	TerminateErrorCode() TerminateErrorCode
}

type DefaultTerminateError struct {
	code   TerminateErrorCode
	reason string
}

func (err DefaultTerminateError) Error() string {
	return err.reason
}

func (err DefaultTerminateError) TerminateErrorCode() TerminateErrorCode {
	return err.code
}
