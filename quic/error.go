package quic

import (
	"github.com/quic-go/quic-go"
)

type TransportError = quic.TransportError

// struct {
// 	Remote       bool
// 	FrameType    uint64
// 	ErrorCode    TransportErrorCode
// 	ErrorMessage string
// 	Err          error // only set for local errors, sometimes
// }

// func (e *TransportError) Error() string        { return "quic: " + e.Err.Error() }
// func (e *TransportError) Is(target error) bool { return target == net.ErrClosed }
// func (e *TransportError) Unwrap() error        { return e.Err }

type ApplicationError = quic.ApplicationError

// struct {
// 	Remote       bool
// 	ErrorCode    ApplicationErrorCode
// 	ErrorMessage string
// 	Err          error
// }

// func (e *ApplicationError) Error() string        { return "quic: " + e.Err.Error() }
// func (e *ApplicationError) Is(target error) bool { return target == net.ErrClosed }
// func (e *ApplicationError) Unwrap() error        { return e.Err }

type VersionNegotiationError = quic.VersionNegotiationError

// struct {
// 	Ours   []Version
// 	Theirs []Version
// 	Err    error
// }

// func (e *VersionNegotiationError) Error() string        { return "quic: " + e.Err.Error() }
// func (e *VersionNegotiationError) Is(target error) bool { return target == net.ErrClosed }

type StatelessResetError = quic.StatelessResetError

// struct {
// 	// Token [16]byte
// 	Err error
// }

// func (e *StatelessResetError) Error() string        { return "quic: " + e.Err.Error() }
// func (e *StatelessResetError) Is(target error) bool { return target == net.ErrClosed }
// func (e *StatelessResetError) Timeout() bool        { return false }
// func (e *StatelessResetError) Temporary() bool      { return true }

type IdleTimeoutError = quic.IdleTimeoutError

// struct {
// 	Err error
// }

// func (e *IdleTimeoutError) Timeout() bool        { return true }
// func (e *IdleTimeoutError) Temporary() bool      { return false }
// func (e *IdleTimeoutError) Error() string        { return e.Err.Error() }
// func (e *IdleTimeoutError) Is(target error) bool { return target == net.ErrClosed }

type HandshakeTimeoutError = quic.HandshakeTimeoutError

// struct {
// 	Err error
// }

// func (e *HandshakeTimeoutError) Timeout() bool        { return true }
// func (e *HandshakeTimeoutError) Temporary() bool      { return false }
// func (e *HandshakeTimeoutError) Error() string        { return e.Err.Error() }
// func (e *HandshakeTimeoutError) Is(target error) bool { return target == net.ErrClosed }

type (
	TransportErrorCode   = quic.TransportErrorCode
	ApplicationErrorCode = quic.ApplicationErrorCode
	StreamErrorCode      = quic.StreamErrorCode
)

const (
	NoError                   TransportErrorCode = quic.NoError
	InternalError             TransportErrorCode = quic.InternalError
	ConnectionRefused         TransportErrorCode = quic.ConnectionRefused
	FlowControlError          TransportErrorCode = quic.FlowControlError
	StreamLimitError          TransportErrorCode = quic.StreamLimitError
	StreamStateError          TransportErrorCode = quic.StreamStateError
	FinalSizeError            TransportErrorCode = quic.FinalSizeError
	FrameEncodingError        TransportErrorCode = quic.FrameEncodingError
	TransportParameterError   TransportErrorCode = quic.TransportParameterError
	ConnectionIDLimitError    TransportErrorCode = quic.ConnectionIDLimitError
	ProtocolViolation         TransportErrorCode = quic.ProtocolViolation
	InvalidToken              TransportErrorCode = quic.InvalidToken
	ApplicationErrorErrorCode TransportErrorCode = quic.ApplicationErrorErrorCode
	CryptoBufferExceeded      TransportErrorCode = quic.CryptoBufferExceeded
	KeyUpdateError            TransportErrorCode = quic.KeyUpdateError
	AEADLimitReached          TransportErrorCode = quic.AEADLimitReached
	NoViablePathError         TransportErrorCode = quic.NoViablePathError
)

// A StreamError is used for Stream.CancelRead and Stream.CancelWrite.
// It is also returned from Stream.Read and Stream.Write if the peer canceled reading or writing.
type StreamError = quic.StreamError

// struct {
// 	StreamID  StreamID
// 	ErrorCode StreamErrorCode
// 	Remote    bool
// 	Err       error
// }

// func (e *StreamError) Is(target error) bool {
// 	_, ok := target.(*StreamError)
// 	return ok
// }

// func (e *StreamError) Error() string { return "quic: " + e.Err.Error() }

// DatagramTooLargeError is returned from Connection.SendDatagram if the payload is too large to be sent.
type DatagramTooLargeError = quic.DatagramTooLargeError

// struct {
// 	MaxDatagramPayloadSize int64
// 	Err                    error
// }

// func (e *DatagramTooLargeError) Is(target error) bool {
// 	_, ok := target.(*DatagramTooLargeError)
// 	return ok
// }

// func (e *DatagramTooLargeError) Error() string { return "quic: " + e.Err.Error() }
