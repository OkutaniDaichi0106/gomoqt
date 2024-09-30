package moqtransport

import "github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"

/*
 * Announce Error
 */

var (
	ErrAnnounceFailed = DefaultAnnounceError{
		reason: "internal error",
		code:   moqtmessage.ANNOUNCE_INTERNAL_ERROR,
	}
	ErrDuplicateTrackNamespace = DefaultAnnounceError{
		reason: "duplicate track namespace",
		code:   moqtmessage.DUPLICATE_TRACK_NAMESPACE,
	}
)

type AnnounceError interface {
	error
	Code() moqtmessage.AnnounceErrorCode
}

type DefaultAnnounceError struct {
	reason string
	code   moqtmessage.AnnounceErrorCode
}

func (err DefaultAnnounceError) Error() string {
	return err.reason
}

func (err DefaultAnnounceError) Code() moqtmessage.AnnounceErrorCode {
	return err.code
}

/*
 * Subscribe Error
 */
var (
	ErrSubscribeFailed = DefaultSubscribeError{
		code:   moqtmessage.SUBSCRIBE_INTERNAL_ERROR,
		reason: "internal error",
	}

	ErrDefaultInvalidRange = DefaultSubscribeError{
		code:   moqtmessage.INVALID_RANGE,
		reason: "invalid range",
	}
)

func GetSubscribeError(message moqtmessage.SubscribeErrorMessage) SubscribeError {
	if message.Code == moqtmessage.RETRY_TRACK_ALIAS {
		return RetryTrackAliasError{
			reason:     message.Reason,
			trackAlias: message.TrackAlias,
		}
	}

	return DefaultSubscribeError{
		code:   message.Code,
		reason: message.Reason,
	}
}

type SubscribeError interface {
	error
	Code() moqtmessage.SubscribeErrorCode
}

type DefaultSubscribeError struct {
	code   moqtmessage.SubscribeErrorCode
	reason string
}

func (err DefaultSubscribeError) Error() string {
	return err.reason
}

func (err DefaultSubscribeError) Code() moqtmessage.SubscribeErrorCode {
	return err.code
}

type RetryTrackAliasError struct {
	//code       moqtmessage.SubscribeErrorCode
	reason     string
	trackAlias moqtmessage.TrackAlias
}

func (err RetryTrackAliasError) Error() string {
	return err.reason
}

func (err RetryTrackAliasError) Code() moqtmessage.SubscribeErrorCode {
	return moqtmessage.RETRY_TRACK_ALIAS
}

func (err RetryTrackAliasError) TrackAlias() moqtmessage.TrackAlias {
	return err.trackAlias
}

/*
 *
 */
type SubscribeDoneStatus interface {
	Reason() string
	Code() moqtmessage.SubscribeDoneStatusCode
}

type SubscribeDoneError interface {
	error
	Code() moqtmessage.SubscribeDoneStatusCode
}

var (
	ErrSubscribeDoneInternalError = DefaultSubscribeDoneError{
		code:   moqtmessage.SUBSCRIBE_DONE_INTERNAL_ERROR,
		reason: "internal error",
	}
	ErrSubscribeUnauthorized = DefaultSubscribeDoneError{
		code:   moqtmessage.SUBSCRIBE_DONE_UNAUTHORIZED,
		reason: "unauthorized",
	}
	ErrSubscribeExpired = DefaultSubscribeDoneError{
		code:   moqtmessage.SUBSCRIBE_DONE_EXPIRED,
		reason: "expired",
	}
)

type DefaultSubscribeDoneError struct {
	code   moqtmessage.SubscribeDoneStatusCode
	reason string
}

func (err DefaultSubscribeDoneError) Error() string {
	return err.reason
}

func (err DefaultSubscribeDoneError) Code() moqtmessage.SubscribeDoneStatusCode {
	return err.code
}

/***/

type AnnounceCancelError interface {
	Code() moqtmessage.AnnounceCancelCode
	Reason() string
}

var _ AnnounceCancelError = (*DefaultAnnounceCancelError)(nil)

type DefaultAnnounceCancelError struct {
	code   moqtmessage.AnnounceCancelCode
	reason string
}

func (cancel DefaultAnnounceCancelError) Code() moqtmessage.AnnounceCancelCode {
	return cancel.code
}

func (cancel DefaultAnnounceCancelError) Reason() string {
	return cancel.reason
}

type SubscribeNamespaceError interface {
	error
	Code() uint64
}

type DefaultSubscribeNamespaceError struct {
	code   moqtmessage.SubscribeNamespaceErrorCode
	reason string
}

func (err DefaultSubscribeNamespaceError) Error() string {
	return err.reason
}

func (err DefaultSubscribeNamespaceError) Code() moqtmessage.SubscribeNamespaceErrorCode {
	return err.code
}
