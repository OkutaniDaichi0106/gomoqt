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
	ErrSubscribeFailed = SubscribeInternalError{}
	ErrInvalidRange    = SubscribeInvalidRange{}
)

type SubscribeError interface {
	error
	Code() moqtmessage.SubscribeErrorCode
}

type SubscribeInternalError struct {
}

func (SubscribeInternalError) Error() string {
	return "internal error"
}

func (SubscribeInternalError) Code() moqtmessage.SubscribeErrorCode {
	return moqtmessage.SUBSCRIBE_INTERNAL_ERROR
}

type SubscribeInvalidRange struct {
}

func (SubscribeInvalidRange) Error() string {
	return "duplicate track namespace"
}

func (SubscribeInvalidRange) Code() moqtmessage.SubscribeErrorCode {
	return moqtmessage.INVALID_RANGE
}
