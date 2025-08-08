package internal

import (
	"net"
)

type TransportError struct {
	Remote       bool
	FrameType    uint64
	ErrorCode    TransportErrorCode
	ErrorMessage string
	Err          error // only set for local errors, sometimes
}

func (e *TransportError) Error() string        { return "quic: " + e.Err.Error() }
func (e *TransportError) Is(target error) bool { return target == net.ErrClosed }
func (e *TransportError) Unwrap() error        { return e.Err }

type ApplicationError struct {
	Remote       bool
	ErrorCode    ApplicationErrorCode
	ErrorMessage string
	Err          error
}

func (e *ApplicationError) Error() string        { return "quic: " + e.Err.Error() }
func (e *ApplicationError) Is(target error) bool { return target == net.ErrClosed }
func (e *ApplicationError) Unwrap() error        { return e.Err }

type VersionNegotiationError struct {
	Ours   []Version
	Theirs []Version
	Err    error
}

func (e *VersionNegotiationError) Error() string        { return "quic: " + e.Err.Error() }
func (e *VersionNegotiationError) Is(target error) bool { return target == net.ErrClosed }

type StatelessResetError struct {
	// Token [16]byte
	Err error
}

func (e *StatelessResetError) Error() string        { return "quic: " + e.Err.Error() }
func (e *StatelessResetError) Is(target error) bool { return target == net.ErrClosed }
func (e *StatelessResetError) Timeout() bool        { return false }
func (e *StatelessResetError) Temporary() bool      { return true }

type IdleTimeoutError struct {
	Err error
}

func (e *IdleTimeoutError) Timeout() bool        { return true }
func (e *IdleTimeoutError) Temporary() bool      { return false }
func (e *IdleTimeoutError) Error() string        { return e.Err.Error() }
func (e *IdleTimeoutError) Is(target error) bool { return target == net.ErrClosed }

type HandshakeTimeoutError struct {
	Err error
}

func (e *HandshakeTimeoutError) Timeout() bool        { return true }
func (e *HandshakeTimeoutError) Temporary() bool      { return false }
func (e *HandshakeTimeoutError) Error() string        { return e.Err.Error() }
func (e *HandshakeTimeoutError) Is(target error) bool { return target == net.ErrClosed }

type (
	TransportErrorCode   uint64
	ApplicationErrorCode uint64
	StreamErrorCode      uint64
)

// const (
// 	NoError                   TransportErrorCode = 0x0
// 	InternalError             TransportErrorCode = 0x1
// 	ConnectionRefused         TransportErrorCode = 0x2
// 	FlowControlError          TransportErrorCode = 0x3
// 	StreamLimitError          TransportErrorCode = 0x4
// 	StreamStateError          TransportErrorCode = 0x5
// 	FinalSizeError            TransportErrorCode = 0x6
// 	FrameEncodingError        TransportErrorCode = 0x7
// 	TransportParameterError   TransportErrorCode = 0x8
// 	ConnectionIDLimitError    TransportErrorCode = 0x9
// 	ProtocolViolation         TransportErrorCode = 0xA
// 	InvalidToken              TransportErrorCode = 0xB
// 	ApplicationErrorErrorCode TransportErrorCode = 0xC
// 	CryptoBufferExceeded      TransportErrorCode = 0xD
// 	KeyUpdateError            TransportErrorCode = 0xE
// 	AEADLimitReached          TransportErrorCode = 0xF
// 	NoViablePathError         TransportErrorCode = 0x10
// )

// A StreamError is used for Stream.CancelRead and Stream.CancelWrite.
// It is also returned from Stream.Read and Stream.Write if the peer canceled reading or writing.
type StreamError struct {
	StreamID  StreamID
	ErrorCode StreamErrorCode
	Remote    bool
	Err       error
}

func (e *StreamError) Is(target error) bool {
	_, ok := target.(*StreamError)
	return ok
}

func (e *StreamError) Error() string { return "quic: " + e.Err.Error() }

// DatagramTooLargeError is returned from Connection.SendDatagram if the payload is too large to be sent.
type DatagramTooLargeError struct {
	MaxDatagramPayloadSize int64
	Err                    error
}

func (e *DatagramTooLargeError) Is(target error) bool {
	_, ok := target.(*DatagramTooLargeError)
	return ok
}

func (e *DatagramTooLargeError) Error() string { return "quic: " + e.Err.Error() }
