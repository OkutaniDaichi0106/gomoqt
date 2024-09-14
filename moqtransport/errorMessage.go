package moqtransport

import (
	"errors"
)

type TerminateErrorCode int
type TerminateError struct {
}

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
	TERMINATION_NO_ERROR                  AnnounceErrorCode = 0x0
	TERMINATION_INTERNAL_ERROR            AnnounceErrorCode = 0x1
	TERMINATION_UNAUTHORIZED              AnnounceErrorCode = 0x2
	TERMINATION_PROTOCOL_VIOLATION        AnnounceErrorCode = 0x3
	TERMINATION_DUPLICATE_TRACK_ALIAS     AnnounceErrorCode = 0x4
	TERMINATION_PARAMETER_LENGTH_MISMATCH AnnounceErrorCode = 0x5
	TERMINATION_GOAWAY_TIMEOUT            AnnounceErrorCode = 0x6
)

var ErrProtocolViolation = errors.New("protocol violation")
var ErrInvalidFilter = errors.New("invalid filter type")
