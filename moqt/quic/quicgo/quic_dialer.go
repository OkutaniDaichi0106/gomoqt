package quicgo

import (
	"context"
	"crypto/tls"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo "github.com/quic-go/quic-go"
)

func Dial(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *quic.Config) (quic.Connection, error) {
	conn, err := quicgo.DialAddrEarly(ctx, addr, tlsConfig, (*quicgo.Config)(quicConfig))
	if err != nil {
		return nil, err
	}

	return WrapConnection(conn), nil
}
