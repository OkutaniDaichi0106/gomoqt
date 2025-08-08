package quic

import (
	"github.com/OkutaniDaichi0106/gomoqt/quic/internal"
)

type TransportError = internal.TransportError

type ApplicationError = internal.ApplicationError

type VersionNegotiationError = internal.VersionNegotiationError

type StatelessResetError = internal.StatelessResetError

type IdleTimeoutError = internal.IdleTimeoutError

type HandshakeTimeoutError = internal.HandshakeTimeoutError

type (
	TransportErrorCode   = internal.TransportErrorCode
	ApplicationErrorCode = internal.ApplicationErrorCode
	StreamErrorCode      = internal.StreamErrorCode
)

const (
	NoError                   TransportErrorCode = 0x0
	InternalError             TransportErrorCode = 0x1
	ConnectionRefused         TransportErrorCode = 0x2
	FlowControlError          TransportErrorCode = 0x3
	StreamLimitError          TransportErrorCode = 0x4
	StreamStateError          TransportErrorCode = 0x5
	FinalSizeError            TransportErrorCode = 0x6
	FrameEncodingError        TransportErrorCode = 0x7
	TransportParameterError   TransportErrorCode = 0x8
	ConnectionIDLimitError    TransportErrorCode = 0x9
	ProtocolViolation         TransportErrorCode = 0xA
	InvalidToken              TransportErrorCode = 0xB
	ApplicationErrorErrorCode TransportErrorCode = 0xC
	CryptoBufferExceeded      TransportErrorCode = 0xD
	KeyUpdateError            TransportErrorCode = 0xE
	AEADLimitReached          TransportErrorCode = 0xF
	NoViablePathError         TransportErrorCode = 0x10
)

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
