package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

/*
 * Stream Error
 */

const (
	stream_internal_error moq.StreamErrorCode = 0x00
	invalid_stream_type   moq.StreamErrorCode = 0x10 // TODO: See spec
)

type defaultStreamError struct {
	code   moq.StreamErrorCode
	reason string
}

func (err defaultStreamError) Error() string {
	return err.reason
}

func (err defaultStreamError) StreamErrorCode() moq.StreamErrorCode {
	return err.code
}

var (
	ErrInvalidStreamType = defaultStreamError{
		code:   invalid_stream_type,
		reason: "invalid stream type",
	}
)

/*
 * Announce Errors
 */
type AnnounceErrorCode uint32

const (
	announce_internal_error            AnnounceErrorCode = 0x0
	announce_duplicate_track_namespace AnnounceErrorCode = 0x1
)

var (
	ErrDuplicatedTrackNamespace = defaultAnnounceError{
		reason: "duplicate track namespace",
		code:   announce_duplicate_track_namespace,
	}
)

type AnnounceError interface {
	error
	AnnounceErrorCode() AnnounceErrorCode
}

type defaultAnnounceError struct {
	reason string
	code   AnnounceErrorCode
}

func (err defaultAnnounceError) Error() string {
	return err.reason
}

func (err defaultAnnounceError) AnnounceErrorCode() AnnounceErrorCode {
	return err.code
}

/*
 * Internal Error
 */
var _ moq.StreamError = (*internalError)(nil)
var _ AnnounceError = (*internalError)(nil)
var _ SubscribeError = (*internalError)(nil)
var _ SubscribeDoneError = (*internalError)(nil)
var _ TerminateError = (*internalError)(nil)
var _ InfoError = (*internalError)(nil)

type internalError struct{}

func (internalError) Error() string {
	return "internal error"
}

func (internalError) AnnounceErrorCode() AnnounceErrorCode {
	return announce_internal_error
}

func (internalError) SubscribeErrorCode() SubscribeErrorCode {
	return subscribe_internal_error
}

func (internalError) SubscribeDoneErrorCode() SubscribeDoneStatusCode {
	return subscribe_done_internal_error
}

func (internalError) TerminateErrorCode() TerminateErrorCode {
	return terminate_internal_error
}

func (internalError) StreamErrorCode() moq.StreamErrorCode {
	return stream_internal_error
}

func (internalError) FetchErrorCode() FetchErrorCode {
	return fetch_internal_error
}

func (internalError) InfoErrorCode() InfoErrorCode {
	return info_internal_error
}

var ErrInternalError = internalError{}

/*
 * Unauthorized Error
 */
var _ SubscribeError = (*unauthorizedError)(nil)
var _ SubscribeDoneError = (*unauthorizedError)(nil)
var _ TerminateError = (*unauthorizedError)(nil)

type unauthorizedError struct{}

func (unauthorizedError) Error() string {
	return "internal error"
}

func (unauthorizedError) SubscribeErrorCode() SubscribeErrorCode {
	return subscribe_unauthorized
}

func (unauthorizedError) SubscribeDoneErrorCode() SubscribeDoneStatusCode {
	return subscribe_done_unauthorized
}

func (unauthorizedError) TerminateErrorCode() TerminateErrorCode {
	return terminate_unauthorized
}

var ErrUnauthorizedError = internalError{}

/*
 * Subscribe Errors
 */
type SubscribeErrorCode uint32

const (
	subscribe_internal_error       SubscribeErrorCode = 0x00
	subscribe_invlid_range         SubscribeErrorCode = 0x01
	subscribe_track_does_not_exist SubscribeErrorCode = 0x03
	subscribe_unauthorized         SubscribeErrorCode = 0x04
	subscribe_timeout              SubscribeErrorCode = 0x05
)

type SubscribeError interface {
	error
	SubscribeErrorCode() SubscribeErrorCode
}

var _ SubscribeError = (*defaultSubscribeError)(nil)

type defaultSubscribeError struct {
	code   SubscribeErrorCode
	reason string
}

func (err defaultSubscribeError) Error() string {
	return err.reason
}

func (err defaultSubscribeError) SubscribeErrorCode() SubscribeErrorCode {
	return err.code
}

var (
	ErrInvalidRange = defaultSubscribeError{
		code:   subscribe_invlid_range,
		reason: "invalid range",
	}

	ErrSubscribeTimeout = defaultSubscribeError{
		code:   subscribe_timeout,
		reason: "time out",
	}
)

/*
 *
 */
type SubscribeDoneStatusCode uint32

const (
	subscribed_done_unsubscribed      SubscribeDoneStatusCode = 0x0
	subscribe_done_internal_error     SubscribeDoneStatusCode = 0x1
	subscribe_done_unauthorized       SubscribeDoneStatusCode = 0x2
	subscribe_done_track_ended        SubscribeDoneStatusCode = 0x3
	subscribe_done_subscription_ended SubscribeDoneStatusCode = 0x4
	subscribe_done_going_away         SubscribeDoneStatusCode = 0x5
	subscribe_done_expired            SubscribeDoneStatusCode = 0x6
)

type SubscribeDoneError interface {
	error
	SubscribeDoneErrorCode() SubscribeDoneStatusCode
}

var (
	ErrSubscribeExpired = defaultSubscribeDoneError{
		code:   subscribe_done_expired,
		reason: "expired",
	}
)

/*
 * Subscribe Done Error
 */
var _ SubscribeDoneError = (*defaultSubscribeDoneError)(nil)

type defaultSubscribeDoneError struct {
	code   SubscribeDoneStatusCode
	reason string
}

func (err defaultSubscribeDoneError) Error() string {
	return err.reason
}

func (err defaultSubscribeDoneError) Reason() string {
	return err.reason
}

func (err defaultSubscribeDoneError) SubscribeDoneErrorCode() SubscribeDoneStatusCode {
	return err.code
}

/*
 * Subscribe Done Status
 */
type SubscribeDoneStatus interface {
	Reason() string
	Code() SubscribeDoneStatusCode
}

var _ SubscribeDoneStatus = (*DefaultSubscribeDoneStatus)(nil)

var (
	StatusUnsubscribed = DefaultSubscribeDoneStatus{
		code:   subscribed_done_unsubscribed,
		reason: "unsubscribed",
	}
	StatusEndedTrack = DefaultSubscribeDoneStatus{
		code:   subscribe_done_track_ended,
		reason: "track ended",
	}
	StatusEndedSubscription = DefaultSubscribeDoneStatus{
		code:   subscribe_done_subscription_ended,
		reason: "subsription ended",
	}
	StatusGoingAway = DefaultSubscribeDoneStatus{
		code:   subscribe_done_going_away,
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

/*
 * Info Errors
 */
type InfoErrorCode int

const (
	info_internal_error       InfoErrorCode = 0x00
	info_track_does_not_exist InfoErrorCode = 0x01
)

type InfoError interface {
	error
	InfoErrorCode() InfoErrorCode
}

var _ InfoError = (*defaultInfoError)(nil)

type defaultInfoError struct {
	code   InfoErrorCode
	reason string
}

func (err defaultInfoError) Error() string {
	return err.reason
}

func (err defaultInfoError) InfoErrorCode() InfoErrorCode {
	return err.code
}

var (
	ErrTrackDoesNotExist = trackNotFoundError{}
)

/*
 * Track Not Found Error
 */
var _ SubscribeError = (*trackNotFoundError)(nil)
var _ InfoError = (*trackNotFoundError)(nil)

type trackNotFoundError struct{}

func (trackNotFoundError) Error() string {
	return "track does not exist"
}
func (trackNotFoundError) SubscribeErrorCode() SubscribeErrorCode {
	return subscribe_track_does_not_exist
}
func (trackNotFoundError) InfoErrorCode() InfoErrorCode {
	return info_track_does_not_exist
}

/*
 * Fetch Errors
 */
type FetchErrorCode int

type FetchError interface {
	error
	FetchErrorCode() FetchErrorCode
}

var _ FetchError = (*defaultFetchError)(nil)

type defaultFetchError struct {
	code   FetchErrorCode
	reason string
}

func (err defaultFetchError) Error() string {
	return err.reason
}

func (err defaultFetchError) FetchErrorCode() FetchErrorCode {
	return err.code
}

const (
	fetch_internal_error FetchErrorCode = 0x0
	fetch_no_group       FetchErrorCode = 0x1
	fetch_invalid_offset FetchErrorCode = 0x2
)

var (
	ErrNoGroup = defaultFetchError{
		code:   fetch_no_group,
		reason: "no group",
	}

	ErrInvalidOffset = defaultFetchError{
		code:   fetch_invalid_offset,
		reason: "invalid offset",
	}
)

/*
 * Terminate Error
 */
type TerminateErrorCode int

const (
	terminate_no_error                  TerminateErrorCode = 0x0
	terminate_internal_error            TerminateErrorCode = 0x1
	terminate_unauthorized              TerminateErrorCode = 0x2
	terminate_protocol_violation        TerminateErrorCode = 0x3
	terminate_duplicate_track_alias     TerminateErrorCode = 0x4
	terminate_parameter_length_mismatch TerminateErrorCode = 0x5
	terminate_too_many_subscribes       TerminateErrorCode = 0x6
	terminate_goaway_timeout            TerminateErrorCode = 0x10
)

type TerminateError interface {
	error
	TerminateErrorCode() TerminateErrorCode
}

var _ TerminateError = (*defaultTerminateError)(nil)

type defaultTerminateError struct {
	code   TerminateErrorCode
	reason string
}

func (err defaultTerminateError) Error() string {
	return err.reason
}

func (err defaultTerminateError) TerminateErrorCode() TerminateErrorCode {
	return err.code
}

var (
	NoErrTerminate = defaultTerminateError{
		code:   terminate_no_error,
		reason: "no error",
	}

	ErrProtocolViolation = defaultTerminateError{
		code:   terminate_protocol_violation,
		reason: "protocol violation",
	}
	ErrDuplicatedTrackAlias = defaultTerminateError{
		code:   terminate_duplicate_track_alias,
		reason: "duplicate track alias",
	}
	ErrParameterLengthMismatch = defaultTerminateError{
		code:   terminate_parameter_length_mismatch,
		reason: "parameter length mismatch",
	}
	ErrTooManySubscribes = defaultTerminateError{
		code:   terminate_too_many_subscribes,
		reason: "too many subscribes",
	}
	ErrGoAwayTimeout = defaultTerminateError{
		code:   terminate_goaway_timeout,
		reason: "goaway timeout",
	}
)

var _ TerminateError = (*ErrorWithGoAway)(nil)

// TODO:
type ErrorWithGoAway struct {
	TerminateError
	NewSessionURI string
	Timeout       time.Duration
}
