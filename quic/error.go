package quic

import (
	"github.com/quic-go/quic-go"
)

type TransportError = quic.TransportError

type ApplicationError = quic.ApplicationError

type VersionNegotiationError = quic.VersionNegotiationError

type StatelessResetError = quic.StatelessResetError

type IdleTimeoutError = quic.IdleTimeoutError

type HandshakeTimeoutError = quic.HandshakeTimeoutError

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

// DatagramTooLargeError is returned from Connection.SendDatagram if the payload is too large to be sent.
type DatagramTooLargeError = quic.DatagramTooLargeError
