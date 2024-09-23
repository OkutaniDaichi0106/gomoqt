package moqtransport

type TerminateErrorCode int

var (
	NoTerminateErr             TerminateNoError
	ErrInternalError           TerminateInternalError
	ErrUnauthorized            terminateUnauthorized
	ErrProtocolViolation       TerminateProtocolViolation
	ErrDuplicatedTrackAlias    terminateDuplicateTrackAlias
	ErrParameterLengthMismatch terminateParameterLengthMismatch
	ErrTooManySubscribes       terminateTooManySubscribes
	ErrGoAwayTimeout           terminateGoAwayTimeout
)

/*
 * Error codes and status codes for termination of the session
 *
 * The following error codes and status codes are defined in the official document
 * NO_ERROR
 * INTERNAL_ERROR
 * UNAUTHORIZED
 * PROTOCOL_VIOLATION
 * DUPLICATE_TRACK_ALIAS
 * PARAMETER_LENGTH_MISMATCH
 * GOAWAY_TIMEOUT
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
	Code() TerminateErrorCode
}

type TerminateNoError struct{}

func (TerminateNoError) Error() string {
	return "no error"
}

func (TerminateNoError) Code() TerminateErrorCode {
	return TERMINATE_NO_ERROR
}

type TerminateInternalError struct{}

func (TerminateInternalError) Error() string {
	return "internal error"
}

func (TerminateInternalError) Code() TerminateErrorCode {
	return TERMINATE_INTERNAL_ERROR
}

type terminateUnauthorized struct{}

func (terminateUnauthorized) Error() string {
	return "unauthorized"
}

func (terminateUnauthorized) Code() TerminateErrorCode {
	return TERMINATE_UNAUTHORIZED
}

type TerminateProtocolViolation struct{}

func (TerminateProtocolViolation) Error() string {
	return "protocol violation"
}

func (TerminateProtocolViolation) Code() TerminateErrorCode {
	return TERMINATE_PROTOCOL_VIOLATION
}

type terminateDuplicateTrackAlias struct{}

func (terminateDuplicateTrackAlias) Error() string {
	return "duplicate track alias"
}

func (terminateDuplicateTrackAlias) Code() TerminateErrorCode {
	return TERMINATE_DUPLICATE_TRACK_ALIAS
}

type terminateParameterLengthMismatch struct{}

func (terminateParameterLengthMismatch) Error() string {
	return "parameter length mismatch"
}

func (terminateParameterLengthMismatch) Code() TerminateErrorCode {
	return TERMINATE_PARAMETER_LENGTH_MISMATCH
}

type terminateTooManySubscribes struct{}

func (terminateTooManySubscribes) Error() string {
	return "too many subscribes"
}

func (terminateTooManySubscribes) Code() TerminateErrorCode {
	return TERMINATE_TOO_MANY_SUBSCRIBES
}

type terminateGoAwayTimeout struct{}

func (terminateGoAwayTimeout) Error() string {
	return "goaway timeout"
}

func (terminateGoAwayTimeout) Code() TerminateErrorCode {
	return TERMINATE_GOAWAY_TIMEOUT
}
