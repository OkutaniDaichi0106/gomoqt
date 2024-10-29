package moqt

/*
 * Announce Error
 */
type AnnounceErrorCode uint32

const (
	announce_internal_error            AnnounceErrorCode = 0x0
	announce_duplicate_track_namespace AnnounceErrorCode = 0x1
)

var (
	ErrDuplicatedTrackNamespace = DefaultAnnounceError{
		reason: "duplicate track namespace",
		code:   announce_duplicate_track_namespace,
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
	return announce_internal_error
}

func (InternalError) SubscribeErrorCode() SubscribeErrorCode {
	return subscribe_internal_error
}

func (InternalError) SubscribeDoneErrorCode() SubscribeDoneStatusCode {
	return subscribe_done_internal_error
}

func (InternalError) TerminateErrorCode() TerminateErrorCode {
	return terminate_internal_error
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
	return subscribe_unauthorized
}

func (UnauthorizedError) SubscribeDoneErrorCode() SubscribeDoneStatusCode {
	return subscribe_done_unauthorized
}

func (UnauthorizedError) TerminateErrorCode() TerminateErrorCode {
	return terminate_unauthorized
}

var ErrUnauthorizedError = InternalError{}

/*
 * Subscribe Error
 */
type SubscribeErrorCode uint32

const (
	subscribe_internal_error       SubscribeErrorCode = 0x00
	subscribe_invlid_range         SubscribeErrorCode = 0x01
	subscribe_retry_track_alias    SubscribeErrorCode = 0x02
	subscribe_track_does_not_exist SubscribeErrorCode = 0x03
	subscribe_unauthorized         SubscribeErrorCode = 0x04
	subscribe_timeout              SubscribeErrorCode = 0x05
)

var (
	ErrDefaultInvalidRange = defaultSubscribeError{
		code:   subscribe_invlid_range,
		reason: "invalid range",
	}

	ErrTrackDoesNotExist = defaultSubscribeError{
		code:   subscribe_track_does_not_exist,
		reason: "track does not exist",
	}

	ErrSubscribeTimeout = defaultSubscribeError{
		code:   subscribe_timeout,
		reason: "time out",
	}
)

type SubscribeError interface {
	error
	SubscribeErrorCode() SubscribeErrorCode
}

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

type retryTrackAliasError struct {
	reason     string
	trackAlias TrackAlias
}

func (err retryTrackAliasError) Error() string {
	return err.reason
}

func (err retryTrackAliasError) SubscribeErrorCode() SubscribeErrorCode {
	return subscribe_retry_track_alias
}

func (err retryTrackAliasError) TrackAlias() TrackAlias {
	return err.trackAlias
}

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
 * Interest Error
 */

type InterestErrorCode int

type InterestError interface {
	error
	InterestErrorCode() InterestErrorCode
}

var _ InterestError = (*defaultInterestError)(nil)

type defaultInterestError struct {
	code   InterestErrorCode
	reason string
}

func (err defaultInterestError) Error() string {
	return err.reason
}

func (err defaultInterestError) InterestErrorCode() InterestErrorCode {
	return err.code
}

var (
	ErrTrackNotFound = defaultInterestError{
		code:   0x00,
		reason: "track not found",
	}
)

/*
 *
 */
type TerminateErrorCode int

type TerminateError interface {
	error
	TerminateErrorCode() TerminateErrorCode
}

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

var (
	NoTerminateErr = defaultTerminateError{
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

type TrackAlias uint64
