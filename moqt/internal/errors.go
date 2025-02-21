package internal

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

var (
	ErrInternalError = internalError{}

	ErrUnauthorizedError = unauthorizedError{}

	ErrTrackDoesNotExist = trackDoesNotExistError{}

	ErrDuplicatedTrack = defaultAnnounceError{
		reason: "duplicated track path",
		code:   announce_duplicated_track_path,
	}

	ErrInvalidRange = defaultSubscribeError{
		code:   subscribe_invalid_range,
		reason: "invalid range",
	}

	ErrDuplicatedSubscribeID = defaultSubscribeError{
		code:   subscriber_duplicated_id,
		reason: "duplicated subscribe id",
	}

	// ErrPriorityMismatch = defaultSubscribeError{
	// 	code:   subscribe_priority_mismatch_error,
	// 	reason: "update failed",
	// }

	// ErrGroupOrderMismatch = defaultSubscribeError{
	// 	code:   subscribe_order_mismatch_error,
	// 	reason: "group order mismatch",
	// }

	// TODO:
	// ErrSubscriptionLimitExceeded

	// ErrSubscribeExpired = defaultSubscribeDoneError{
	// 	code:   subscribe_done_expired,
	// 	reason: "expired",
	// }

	NoErrTerminate = defaultTerminateError{
		code:   terminate_no_error,
		reason: "no error",
	}

	ErrProtocolViolation = defaultTerminateError{
		code:   terminate_protocol_violation,
		reason: "protocol violation",
	}

	ErrParameterLengthMismatch = defaultTerminateError{
		code:   terminate_parameter_length_mismatch,
		reason: "parameter length mismatch",
	}

	ErrTooManySubscribes = defaultTerminateError{
		code:   terminate_too_many_subscribes,
		reason: "too many subscribes",
	}

	ErrGroupRejected = defaultGroupError{
		code:   group_send_interrupted,
		reason: "send interrupted",
	}

	ErrGroupOutOfRange = defaultGroupError{
		code:   group_out_of_range,
		reason: "out of range",
	}

	ErrGroupExpired = defaultGroupError{
		code:   group_expires,
		reason: "expires",
	}

	ErrClosedGroup = defaultGroupError{
		code:   group_closed,
		reason: "group is closed",
	}

	// ErrGroupDeliveryTimeout = defaultGroupError{
	// 	code:   group_delivery_timeout,
	// 	reason: "delivery timeout",
	// }

	// ErrDuplicatedGroup = defaultGroupError{
	// 	code:   group_duplicated_group,
	// 	reason: "duplicated group",
	// }
)

// type Error interface {
// 	error
// 	ErrorCode() ErrorCode
// }

/*
 * Session Error
 */

// const (
// 	session_internal_error transport.SessionErrorCode = 0x00
// )

/*
 * Stream Error
 */

const (
	stream_internal_error transport.StreamErrorCode = 0x00
	invalid_stream_type   transport.StreamErrorCode = 0x10 // TODO: See spec
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
const (
	announce_internal_error        protocol.AnnounceErrorCode = 0x0
	announce_duplicated_track_path protocol.AnnounceErrorCode = 0x1
	announce_duplicated_interest   protocol.AnnounceErrorCode = 0x2
)

type AnnounceError interface {
	error
	AnnounceErrorCode() protocol.AnnounceErrorCode
}

type defaultAnnounceError struct {
	reason string
	code   protocol.AnnounceErrorCode
}

func (err defaultAnnounceError) Error() string {
	return err.reason
}

func (err defaultAnnounceError) AnnounceErrorCode() protocol.AnnounceErrorCode {
	return err.code
}

/*
 * Subscribe Errors
 */
const (
	subscribe_internal_error       protocol.SubscribeErrorCode = 0x00
	subscribe_invalid_range        protocol.SubscribeErrorCode = 0x01
	subscriber_duplicated_id       protocol.SubscribeErrorCode = 0x02
	subscribe_track_does_not_exist protocol.SubscribeErrorCode = 0x03
	subscribe_unauthorized         protocol.SubscribeErrorCode = 0x04
	subscribe_timeout              protocol.SubscribeErrorCode = 0x05
	subscribe_update_error         protocol.SubscribeErrorCode = 0x06
	subscribe_closed_track         protocol.SubscribeErrorCode = 0x07
	subscribe_ended_track          protocol.SubscribeErrorCode = 0x08
	// subscribe_priority_mismatch_error protocol.SubscribeErrorCode = 0x07
	// subscribe_order_mismatch_error    protocol.SubscribeErrorCode = 0x08
)

type SubscribeError interface {
	error
	SubscribeErrorCode() protocol.SubscribeErrorCode
}

var _ SubscribeError = (*defaultSubscribeError)(nil)

type defaultSubscribeError struct {
	code   protocol.SubscribeErrorCode
	reason string
}

func (err defaultSubscribeError) Error() string {
	return err.reason
}

func (err defaultSubscribeError) SubscribeErrorCode() protocol.SubscribeErrorCode {
	return err.code
}

/*
 * Info Errors
 */
const (
	info_internal_error       protocol.InfoErrorCode = 0x00
	info_track_does_not_exist protocol.InfoErrorCode = 0x01
)

type InfoError interface {
	error
	InfoErrorCode() protocol.InfoErrorCode
}

var _ InfoError = (*defaultInfoError)(nil)

type defaultInfoError struct {
	code   protocol.InfoErrorCode
	reason string
}

func (err defaultInfoError) Error() string {
	return err.reason
}

func (err defaultInfoError) InfoErrorCode() protocol.InfoErrorCode {
	return err.code
}

/*
 * Terminate Error
 */
const (
	terminate_no_error protocol.TerminateErrorCode = 0x0

	terminate_internal_error     protocol.TerminateErrorCode = 0x1
	terminate_unauthorized       protocol.TerminateErrorCode = 0x2
	terminate_protocol_violation protocol.TerminateErrorCode = 0x3
	// terminate_duplicate_track           protocol.TerminateErrorCode = 0x4
	terminate_parameter_length_mismatch protocol.TerminateErrorCode = 0x5
	terminate_too_many_subscribes       protocol.TerminateErrorCode = 0x6
	terminate_goaway_timeout            protocol.TerminateErrorCode = 0x10
	terminate_handle_timeout            protocol.TerminateErrorCode = 0x11
)

type TerminateError interface {
	error
	TerminateErrorCode() protocol.TerminateErrorCode
}

var _ TerminateError = (*defaultTerminateError)(nil)

type defaultTerminateError struct {
	code   protocol.TerminateErrorCode
	reason string
}

func (err defaultTerminateError) Error() string {
	return err.reason
}

func (err defaultTerminateError) TerminateErrorCode() protocol.TerminateErrorCode {
	return err.code
}

/*
 * Group Error
 */
type GroupError interface {
	GroupErrorCode() protocol.GroupErrorCode
}

const (
	group_internal_error protocol.GroupErrorCode = 0x00

	group_send_interrupted     protocol.GroupErrorCode = 0x01
	group_out_of_range         protocol.GroupErrorCode = 0x02
	group_expires              protocol.GroupErrorCode = 0x03
	group_closed               protocol.GroupErrorCode = 0x04
	group_track_does_not_exist protocol.GroupErrorCode = 0x05

	// group_duplicated_group protocol.GroupErrorCode = 0x10
)

type defaultGroupError struct {
	code   protocol.GroupErrorCode
	reason string
}

func (err defaultGroupError) Error() string {
	return err.reason
}

func (err defaultGroupError) GroupErrorCode() protocol.GroupErrorCode {
	return err.code
}

/*
 * Internal Error
 */
var _ transport.StreamError = (*internalError)(nil)
var _ TerminateError = (*internalError)(nil)
var _ AnnounceError = (*internalError)(nil)
var _ SubscribeError = (*internalError)(nil)
var _ InfoError = (*internalError)(nil)
var _ GroupError = (*internalError)(nil)

type internalError struct{}

func (internalError) Error() string {
	return "internal error"
}

func (internalError) AnnounceErrorCode() protocol.AnnounceErrorCode {
	return announce_internal_error
}

func (internalError) SubscribeErrorCode() protocol.SubscribeErrorCode {
	return subscribe_internal_error
}

func (internalError) TerminateErrorCode() protocol.TerminateErrorCode {
	return terminate_internal_error
}

func (internalError) StreamErrorCode() transport.StreamErrorCode {
	return stream_internal_error
}

func (internalError) InfoErrorCode() protocol.InfoErrorCode {
	return info_internal_error
}

func (internalError) GroupErrorCode() protocol.GroupErrorCode {
	return group_internal_error
}

/*
* Unauthorized Error
 */
var _ SubscribeError = (*unauthorizedError)(nil)
var _ TerminateError = (*unauthorizedError)(nil)

type unauthorizedError struct{}

func (unauthorizedError) Error() string {
	return "unauthorized"
}

func (unauthorizedError) SubscribeErrorCode() protocol.SubscribeErrorCode {
	return subscribe_unauthorized
}

func (unauthorizedError) TerminateErrorCode() protocol.TerminateErrorCode {
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

func (trackDoesNotExistError) SubscribeErrorCode() protocol.SubscribeErrorCode {
	return subscribe_track_does_not_exist
}
func (trackDoesNotExistError) InfoErrorCode() protocol.InfoErrorCode {
	return info_track_does_not_exist
}

func (trackDoesNotExistError) GroupErrorCode() protocol.GroupErrorCode {
	return group_track_does_not_exist
}
