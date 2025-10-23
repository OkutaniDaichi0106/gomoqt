package quic

import (
	"context"
	"crypto/tls"
)

// DialAddrFunc is a function type for establishing a QUIC connection to a remote address.
type DialAddrFunc func(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *Config) (Connection, error)
