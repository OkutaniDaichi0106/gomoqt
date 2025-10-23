package quic

import (
	"context"
	"crypto/tls"
	"net"
)

// ListenAddrFunc is a function type for creating a QUIC listener.
type ListenAddrFunc func(addr string, tlsConfig *tls.Config, quicConfig *Config) (Listener, error)

// Listener accepts incoming QUIC connections.
type Listener interface {
	// Accept waits for and returns the next incoming connection.
	Accept(ctx context.Context) (Connection, error)
	
	// Addr returns the listener's network address.
	Addr() net.Addr
	
	// Close closes the listener and stops accepting new connections.
	Close() error
}
