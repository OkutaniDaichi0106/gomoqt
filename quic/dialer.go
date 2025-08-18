package quic

import (
	"context"
	"crypto/tls"
)

type DialAddrFunc func(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *Config) (Connection, error)
