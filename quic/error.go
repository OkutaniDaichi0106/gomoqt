package quic

import (
	"github.com/quic-go/quic-go"
)

// TransportError represents a QUIC transport layer error.
type TransportError = quic.TransportError

// ApplicationError represents an application-level error in QUIC.
type ApplicationError = quic.ApplicationError

// VersionNegotiationError occurs when version negotiation fails.
type VersionNegotiationError = quic.VersionNegotiationError

// StatelessResetError indicates that a stateless reset was received.
type StatelessResetError = quic.StatelessResetError

// IdleTimeoutError indicates that the connection timed out due to inactivity.
type IdleTimeoutError = quic.IdleTimeoutError

// HandshakeTimeoutError indicates that the handshake did not complete in time.
type HandshakeTimeoutError = quic.HandshakeTimeoutError

// Error codes for QUIC transport, application, and stream operations.
type (
	// TransportErrorCode identifies transport-layer protocol errors.
	TransportErrorCode   = quic.TransportErrorCode
	// ApplicationErrorCode identifies application-defined errors.
	ApplicationErrorCode = quic.ApplicationErrorCode
	// StreamErrorCode identifies stream-specific errors.
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

// DatagramTooLargeError is returned from Connection.SendDatagram if the payload is too large to be sent.
type DatagramTooLargeError = quic.DatagramTooLargeError
