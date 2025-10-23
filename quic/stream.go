package quic

import (
	"context"
	"io"
	"time"

	"github.com/quic-go/quic-go"
)

// Stream is a bidirectional QUIC stream that implements both SendStream and ReceiveStream.
type Stream interface {
	SendStream
	ReceiveStream
	// SetDeadline sets the read and write deadlines.
	SetDeadline(time.Time) error
}

// SendStream is a unidirectional QUIC stream for sending data.
type SendStream interface {
	io.Writer
	io.Closer

	// StreamID returns the stream's unique identifier.
	StreamID() StreamID
	
	// CancelWrite cancels writing with the given error code.
	CancelWrite(StreamErrorCode)

	// SetWriteDeadline sets the deadline for write operations.
	SetWriteDeadline(time.Time) error

	// Context returns the stream's context, canceled when the stream is closed.
	Context() context.Context
}

// ReceiveStream is a unidirectional QUIC stream for receiving data.
type ReceiveStream interface {
	io.Reader

	// StreamID returns the stream's unique identifier.
	StreamID() StreamID

	// CancelRead cancels reading with the given error code.
	CancelRead(StreamErrorCode)

	// SetReadDeadline sets the deadline for read operations.
	SetReadDeadline(time.Time) error
}

// StreamID uniquely identifies a stream within a QUIC connection.
type StreamID = quic.StreamID
