package quic

import (
	"context"
	"net"

	"github.com/quic-go/quic-go"
)

// Connection represents a QUIC connection that can send and receive streams.
// It abstracts the underlying QUIC implementation and provides methods for
// creating bidirectional and unidirectional streams.
type Connection interface {
	// AcceptStream waits for and accepts the next incoming bidirectional stream.
	AcceptStream(ctx context.Context) (Stream, error)

	// AcceptUniStream waits for and accepts the next incoming unidirectional stream.
	AcceptUniStream(ctx context.Context) (ReceiveStream, error)

	// CloseWithError closes the connection with an error code and message.
	CloseWithError(code ApplicationErrorCode, msg string) error

	// ConnectionState returns the current state of the connection.
	ConnectionState() ConnectionState

	// ConnectionStats returns statistics about the connection.
	ConnectionStats() ConnectionStats

	// Context returns the connection's context, which is canceled when the connection is closed.
	Context() context.Context

	// LocalAddr returns the local network address.
	LocalAddr() net.Addr

	// OpenStream opens a new bidirectional stream without blocking.
	OpenStream() (Stream, error)

	// OpenStreamSync opens a new bidirectional stream, blocking until complete.
	OpenStreamSync(ctx context.Context) (Stream, error)

	// OpenUniStream opens a new unidirectional stream without blocking.
	OpenUniStream() (SendStream, error)

	// OpenUniStreamSync opens a new unidirectional stream, blocking until complete.
	OpenUniStreamSync(ctx context.Context) (str SendStream, err error)

	// RemoteAddr returns the remote network address.
	RemoteAddr() net.Addr
}

// ConnectionState holds information about the QUIC connection state.
type ConnectionState = quic.ConnectionState

type ConnectionStats = quic.ConnectionStats
