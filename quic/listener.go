package quic

import (
	"context"
	"crypto/tls"
	"net"
)

type ListenAddrFunc func(addr string, tlsConfig *tls.Config, quicConfig *Config) (Listener, error)

type Listener interface {
	Accept(ctx context.Context) (Connection, error)
	Addr() net.Addr
	Close() error
}
