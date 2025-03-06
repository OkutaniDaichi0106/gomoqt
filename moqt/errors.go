package moqt

import (
	"errors"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
)

var (
	ErrInternalError = &Error{internal.ErrInternalError}

	ErrUnauthorizedError = &Error{internal.ErrUnauthorizedError} // TODO: Use this error

	ErrTrackDoesNotExist = &Error{internal.ErrTrackDoesNotExist}

	ErrDuplicatedTrack = &Error{internal.ErrDuplicatedTrack}

	ErrInvalidRange = &Error{internal.ErrInvalidRange}

	ErrDuplicatedSubscribeID = &Error{internal.ErrDuplicatedSubscribeID} // TODO: Use this error

	ErrEndedTrack = &Error{internal.ErrEndedTrack}

	ErrClosedTrack = &Error{internal.ErrClosedTrack}

	NoErrTerminate = &Error{internal.NoErrTerminate}

	ErrProtocolViolation = &Error{internal.ErrProtocolViolation}

	// ErrParameterLengthMismatch = &Error{internal.ErrParameterLengthMismatch} // TODO: Remove if not used

	// ErrTooManySubscribes = &Error{internal.ErrTooManySubscribes} // TODO: Remove if not used

	ErrGroupRejected = &Error{internal.ErrGroupRejected}

	ErrGroupOutOfRange = &Error{internal.ErrGroupOutOfRange}

	ErrGroupExpired = &Error{internal.ErrGroupExpired}

	// ErrGroupDeliveryTimeout = &Error{internal.ErrGroupDeliveryTimeout} // TODO: Remove if not used

	// ErrDuplicatedGroup = &Error{internal.ErrDuplicatedGroup} // TODO: Remove if not used

	ErrClosedGroup = &Error{internal.ErrClosedGroup}
)

type Error struct {
	internalError error
}

func (e *Error) Error() string {
	return e.internalError.Error()
}

func (e *Error) Unwrap() error {
	return e.internalError
}

func (e *Error) Is(target error) bool {
	return errors.Is(e.internalError, target)
}

func (e *Error) As(target any) bool {
	// Check if the internal error implements the target interface
	switch ptr := target.(type) {
	case *AnnounceError:
		if ae, ok := e.internalError.(internal.AnnounceError); ok {
			*ptr = &defaultAnnounceError{
				reason: ae.Error(),
				code:   AnnounceErrorCode(ae.AnnounceErrorCode()),
			}
			return true
		}
	case *SubscribeError:
		if se, ok := e.internalError.(internal.SubscribeError); ok {
			*ptr = &defaultSubscribeError{
				reason: se.Error(),
				code:   SubscribeErrorCode(se.SubscribeErrorCode()),
			}
			return true
		}
	case *InfoError:
		if ie, ok := e.internalError.(internal.InfoError); ok {
			*ptr = &defaultInfoError{
				reason: ie.Error(),
				code:   InfoErrorCode(ie.InfoErrorCode()),
			}
			return true
		}
	case *GroupError:
		if ge, ok := e.internalError.(internal.GroupError); ok {
			*ptr = &defaultGroupError{
				reason: e.internalError.Error(),
				code:   GroupErrorCode(ge.GroupErrorCode()),
			}
			return true
		}
	}

	// Fallback to standard errors.As behavior
	return errors.As(e.internalError, target)
}

type AnnounceError interface {
	error
	AnnounceErrorCode() AnnounceErrorCode
}

type AnnounceErrorCode uint64

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

type SubscribeError interface {
	error
	SubscribeErrorCode() SubscribeErrorCode
}

type SubscribeErrorCode uint64

type defaultSubscribeError struct {
	reason string
	code   SubscribeErrorCode
}

func (err defaultSubscribeError) Error() string {
	return err.reason
}

func (err defaultSubscribeError) SubscribeErrorCode() SubscribeErrorCode {
	return err.code
}

type InfoError interface {
	error
	InfoErrorCode() InfoErrorCode
}

type InfoErrorCode uint64

type defaultInfoError struct {
	reason string
	code   InfoErrorCode
}

func (err defaultInfoError) Error() string {
	return err.reason
}

func (err defaultInfoError) InfoErrorCode() InfoErrorCode {
	return err.code
}

func NewGroupError(reason string, code GroupErrorCode) GroupError {
	return &defaultGroupError{reason: reason, code: code}
}

type GroupError interface {
	error
	GroupErrorCode() GroupErrorCode
}

type GroupErrorCode uint64

type defaultGroupError struct {
	reason string
	code   GroupErrorCode
}

func (err defaultGroupError) Error() string {
	return err.reason
}

func (err defaultGroupError) GroupErrorCode() GroupErrorCode {
	return err.code
}

type TerminateError interface {
	error
	TerminateErrorCode() TerminateErrorCode
}

type TerminateErrorCode uint64

type defaultTerminateError struct {
	reason string
	code   TerminateErrorCode
}

func (err defaultTerminateError) Error() string {
	return err.reason
}

func (err defaultTerminateError) TerminateErrorCode() TerminateErrorCode {
	return err.code
}
