package quicgo

import (
	"context"
	"crypto/tls"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo_quicgo "github.com/quic-go/quic-go"
)

func Dial(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *quic.Config) (quic.Connection, error) {
	conn, err := quicgo_quicgo.DialAddrEarly(ctx, addr, tlsConfig, (*quicgo_quicgo.Config)(quicConfig))
	return WrapConnection(conn), WrapError(err)
}
