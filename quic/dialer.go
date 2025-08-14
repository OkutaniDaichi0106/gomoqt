package quic

import (
	"context"
	"crypto/tls"
)

type DialAddrEarlyFunc func(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *Config) (Connection, error)
