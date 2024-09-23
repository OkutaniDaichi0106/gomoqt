package moqtransport

import "go-moq/moqtransport/moqtmessage"

/*
 * Announce Error
 */

var (
	ErrAnnounceFailed          = AnnounceInternalError{}
	ErrDuplicateTrackNamespace = AnnounceDuplicateTrackNamespace{}
)

type AnnounceError interface {
	error
	Code() moqtmessage.AnnounceErrorCode
}

type AnnounceInternalError struct {
}

func (AnnounceInternalError) Error() string {
	return "internal error"
}

func (AnnounceInternalError) Code() moqtmessage.AnnounceErrorCode {
	return moqtmessage.ANNOUNCE_INTERNAL_ERROR
}

type AnnounceDuplicateTrackNamespace struct {
}

func (AnnounceDuplicateTrackNamespace) Error() string {
	return "duplicate track namespace"
}

func (AnnounceDuplicateTrackNamespace) Code() moqtmessage.AnnounceErrorCode {
	return moqtmessage.DUPLICATE_TRACK_NAMESPACE
}

/*
 * Subscribe Error
 */
var (
	ErrDefaultSubscribeFailed SubscribeInternalError
	ErrDefaultInvalidRange    InvalidRangeError
	ErrDefaultRetryTrackAlias RetryTrackAlias
	//ErrDefaultDuplicatedSubscribeID DuplicatedSubscribeID
)

type SubscribeError interface {
	error
	Code() moqtmessage.SubscribeErrorCode
	SubscribeID() moqtmessage.SubscribeID
}

type SubscribeInternalError struct {
	subscribeID moqtmessage.SubscribeID
}

func (SubscribeInternalError) Error() string {
	return "internal error"
}

func (SubscribeInternalError) Code() moqtmessage.SubscribeErrorCode {
	return moqtmessage.SUBSCRIBE_INTERNAL_ERROR
}

func (subErr SubscribeInternalError) SubscribeID() moqtmessage.SubscribeID {
	return subErr.subscribeID
}

func (subErr SubscribeInternalError) NewSubscribeID(id moqtmessage.SubscribeID) SubscribeInternalError {
	subErr.subscribeID = id
	return subErr
}

type InvalidRangeError struct {
	subscribeID moqtmessage.SubscribeID
}

func (InvalidRangeError) Error() string {
	return "duplicated track namespace"
}

func (InvalidRangeError) Code() moqtmessage.SubscribeErrorCode {
	return moqtmessage.INVALID_RANGE
}

func (subErr InvalidRangeError) SubscribeID() moqtmessage.SubscribeID {
	return subErr.subscribeID
}

func (subErr InvalidRangeError) NewSubscribeID(id moqtmessage.SubscribeID) InvalidRangeError {
	subErr.subscribeID = id
	return subErr
}

type RetryTrackAlias struct {
	subscribeID   moqtmessage.SubscribeID
	newTrackAlias moqtmessage.TrackAlias
}

func (RetryTrackAlias) Error() string {
	return "retry track alias"
}

func (RetryTrackAlias) Code() moqtmessage.SubscribeErrorCode {
	return moqtmessage.RETRY_TRACK_ALIAS
}

func (subErr RetryTrackAlias) SubscribeID() moqtmessage.SubscribeID {
	return subErr.subscribeID
}

func (subErr RetryTrackAlias) NewTrackAlias(alias moqtmessage.TrackAlias) RetryTrackAlias {
	subErr.newTrackAlias = alias
	return subErr
}

func (subErr RetryTrackAlias) NewSubscribeID(id moqtmessage.SubscribeID) RetryTrackAlias {
	subErr.subscribeID = id
	return subErr
}

// Original Subscribe Error
// type DuplicatedSubscribeID struct {
// 	subscribeID moqtmessage.SubscribeID
// }

// func (DuplicatedSubscribeID) Error() string {
// 	return "dublicated subscribe id"
// }

// func (DuplicatedSubscribeID) Code() moqtmessage.SubscribeErrorCode {
// 	return 0x03
// }

// func (subErr DuplicatedSubscribeID) SubscribeID() moqtmessage.SubscribeID {
// 	return subErr.subscribeID
// }

// func (subErr DuplicatedSubscribeID) NewSubscribeID(id moqtmessage.SubscribeID) DuplicatedSubscribeID {
// 	subErr.subscribeID = id
// 	return subErr
// }

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
	ErrUnexpectedSubscribeDone SubscribeDoneInternalError
	ErrSubscribeUnauthorized   SubscribeDoneUnauthorized
	ErrSubscribeExpired        SubscribeDoneExpired
)

type SubscribeDoneUnsubscribed struct{}

func (SubscribeDoneUnsubscribed) Reason() string {
	return "unsubscribed"
}

func (SubscribeDoneUnsubscribed) Code() moqtmessage.SubscribeDoneStatusCode {
	return moqtmessage.SUBSCRIBE_DONE_UNSUBSCRIBED
}

type SubscribeDoneInternalError struct{}

func (err SubscribeDoneInternalError) Error() string {
	return err.Reason()
}
func (SubscribeDoneInternalError) Reason() string {
	return "unsubscribed"
}

func (SubscribeDoneInternalError) Code() moqtmessage.SubscribeDoneStatusCode {
	return moqtmessage.SUBSCRIBE_DONE_INTERNAL_ERROR
}

type SubscribeDoneUnauthorized struct{}

func (err SubscribeDoneUnauthorized) Error() string {
	return err.Reason()
}

func (SubscribeDoneUnauthorized) Reason() string {
	return "unauthorized"
}

func (SubscribeDoneUnauthorized) Code() moqtmessage.SubscribeDoneStatusCode {
	return moqtmessage.SUBSCRIBE_DONE_UNAUTHORIZED
}

type SubscribeDoneTrackEnded struct{}

func (SubscribeDoneTrackEnded) Reason() string {
	return "track ended"
}

func (SubscribeDoneTrackEnded) Code() moqtmessage.SubscribeDoneStatusCode {
	return moqtmessage.SUBSCRIBE_DONE_TRACK_ENDED
}

type SubscribeDoneSubscriptionEnded struct{}

func (SubscribeDoneSubscriptionEnded) Reason() string {
	return "subscription ended"
}

func (SubscribeDoneSubscriptionEnded) Code() moqtmessage.SubscribeDoneStatusCode {
	return moqtmessage.SUBSCRIBE_DONE_SUBSCRIPTION_ENDED
}

type SubscribeDoneGoingAway struct{}

func (SubscribeDoneGoingAway) Reason() string {
	return "going away"
}
func (SubscribeDoneGoingAway) Code() moqtmessage.SubscribeDoneStatusCode {
	return moqtmessage.SUBSCRIBE_DONE_GOING_AWAY
}

type SubscribeDoneExpired struct{}

func (err SubscribeDoneExpired) Error() string {
	return err.Reason()
}
func (SubscribeDoneExpired) Reason() string {
	return "expired"
}
func (SubscribeDoneExpired) Code() moqtmessage.SubscribeDoneStatusCode {
	return moqtmessage.SUBSCRIBE_DONE_EXPIRED
}
