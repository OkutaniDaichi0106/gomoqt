package quic

import (
	"context"
	"crypto/tls"

	"github.com/OkutaniDaichi0106/gomoqt/quic/quicgo"
)

func DialDefault(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *Config) (Connection, error) {
	return quicgo.DialAddrEarly(ctx, addr, tlsConfig, quicConfig)
}
