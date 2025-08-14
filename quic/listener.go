package quic

import (
	"context"
	"crypto/tls"
	"net"
)

type ListenAddrEarlyFunc func(addr string, tlsConfig *tls.Config, quicConfig *Config) (EarlyListener, error)

type EarlyListener interface {
	Accept(ctx context.Context) (Connection, error)
	Addr() net.Addr
	Close() error
}
