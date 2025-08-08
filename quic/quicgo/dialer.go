package quicgo

import (
	"context"
	"crypto/tls"

	"github.com/OkutaniDaichi0106/gomoqt/quic/internal"
	quicgo_quicgo "github.com/quic-go/quic-go"
)

func DialAddrEarly(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *internal.Config) (internal.Connection, error) {
	conn, err := quicgo_quicgo.DialAddrEarly(ctx, addr, tlsConfig, quicConfig)

	return wrapConnection(conn), wrapError(err)
}
