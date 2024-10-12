package moqtransport

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
