package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

var (
	ErrInternalError = internal.ErrInternalError

	ErrUnauthorizedError = internal.ErrUnauthorizedError

	ErrTrackDoesNotExist = internal.ErrTrackDoesNotExist

	ErrDuplicatedTrack = internal.ErrDuplicatedTrack

	ErrInvalidRange = internal.ErrInvalidRange

	ErrDuplicatedSubscribeID = internal.ErrDuplicatedSubscribeID

	ErrPriorityMismatch = internal.ErrPriorityMismatch

	ErrGroupOrderMismatch = internal.ErrGroupOrderMismatch

	// TODO:
	// ErrSubscriptionLimitExceeded

	ErrSubscribeExpired = internal.ErrSubscribeExpired

	NoErrTerminate = internal.NoErrTerminate

	ErrProtocolViolation = internal.ErrProtocolViolation

	ErrParameterLengthMismatch = internal.ErrParameterLengthMismatch

	ErrTooManySubscribes = internal.ErrTooManySubscribes

	ErrNoGroup = internal.ErrNoGroup

	ErrUnavailabelFrame = internal.ErrUnavailabelFrame

	ErrGroupSendInterrupted = internal.ErrGroupSendInterrupted

	ErrGroupOutOfRange = internal.ErrGroupOutOfRange

	ErrGroupExpires = internal.ErrGroupExpires

	ErrGroupDeliveryTimeout = internal.ErrGroupDeliveryTimeout

	ErrDuplicatedGroup = internal.ErrDuplicatedGroup
)

/*
 * Session Error
 */
// const (
// 	session_internal_error transport.SessionErrorCode = 0x00
// )

/*
 * Stream Error
 */

type StreamErrorCode transport.StreamErrorCode

const (
	stream_internal_error StreamErrorCode = 0x00
	invalid_stream_type   StreamErrorCode = 0x10 // TODO: See spec
)

// type defaultStreamError struct {
// 	code   transport.StreamErrorCode
// 	reason string
// }

// func (err defaultStreamError) Error() string {
// 	return err.reason
// }

// func (err defaultStreamError) StreamErrorCode() transport.StreamErrorCode {
// 	return err.code
// }

/*
 * Announce Errors
 */
type AnnounceErrorCode uint32

const (
	announce_internal_error        AnnounceErrorCode = 0x0
	announce_duplicated_track_path AnnounceErrorCode = 0x1
	announce_duplicated_interest   AnnounceErrorCode = 0x2
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
 * Subscribe Errors
 */
type SubscribeErrorCode uint32

const (
	subscribe_internal_error          SubscribeErrorCode = 0x00
	subscribe_invalid_range           SubscribeErrorCode = 0x01
	subscriber_duplicated_id          SubscribeErrorCode = 0x02
	subscribe_track_does_not_exist    SubscribeErrorCode = 0x03
	subscribe_unauthorized            SubscribeErrorCode = 0x04
	subscribe_timeout                 SubscribeErrorCode = 0x05
	subscribe_update_error            SubscribeErrorCode = 0x06
	subscribe_priority_mismatch_error SubscribeErrorCode = 0x07
	subscribe_order_mismatch_error    SubscribeErrorCode = 0x08
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
	fetch_internal_error    FetchErrorCode = 0x0
	fetch_no_group          FetchErrorCode = 0x1
	fetch_unavailable_frame FetchErrorCode = 0x2
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

/*
 * Group Error
 */
type GroupError interface {
	GroupErrorCode() GroupErrorCode
}

type GroupErrorCode message.GroupErrorCode

const (
	group_internal_error       GroupErrorCode = 0x00
	group_send_interrupted     GroupErrorCode = 0x01
	group_out_of_range         GroupErrorCode = 0x02
	group_expires              GroupErrorCode = 0x03
	group_delivery_timeout     GroupErrorCode = 0x04
	group_track_does_not_exist GroupErrorCode = 0x05

	group_duplicated_group GroupErrorCode = 0x10
)

type defaultGroupError struct {
	code   GroupErrorCode
	reason string
}

func (err defaultGroupError) Error() string {
	return err.reason
}

func (err defaultGroupError) GroupErrorCode() GroupErrorCode {
	return err.code
}

/*
 * Internal Error
 */
var _ transport.StreamError = (*internalError)(nil)
var _ AnnounceError = (*internalError)(nil)
var _ SubscribeError = (*internalError)(nil)
var _ SubscribeDoneError = (*internalError)(nil)
var _ TerminateError = (*internalError)(nil)
var _ InfoError = (*internalError)(nil)
var _ GroupError = (*internalError)(nil)

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

func (internalError) StreamErrorCode() transport.StreamErrorCode {
	return transport.StreamErrorCode(stream_internal_error)
}

func (internalError) FetchErrorCode() FetchErrorCode {
	return fetch_internal_error
}

func (internalError) InfoErrorCode() InfoErrorCode {
	return info_internal_error
}

func (internalError) GroupErrorCode() GroupErrorCode {
	return group_internal_error
}

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

/*
 * Track Does Not Exist Error
 */
var _ SubscribeError = (*trackDoesNotExistError)(nil)
var _ InfoError = (*trackDoesNotExistError)(nil)

type trackDoesNotExistError struct{}

func (trackDoesNotExistError) Error() string {
	return "track does not exist"
}

func (trackDoesNotExistError) SubscribeErrorCode() SubscribeErrorCode {
	return subscribe_track_does_not_exist
}
func (trackDoesNotExistError) InfoErrorCode() InfoErrorCode {
	return info_track_does_not_exist
}

func (trackDoesNotExistError) GroupErrorCode() GroupErrorCode {
	return group_track_does_not_exist
}
