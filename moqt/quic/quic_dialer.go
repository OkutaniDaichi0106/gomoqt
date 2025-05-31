package quic

import (
	"context"
	"crypto/tls"

	quicgo "github.com/quic-go/quic-go"
)

var DialFunc = defaultDialQUICFunc

func defaultDialQUICFunc(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *Config) (Connection, error) {
	conn, err := quicgo.DialAddrEarly(ctx, addr, tlsConfig, (*quicgo.Config)(quicConfig))
	if err != nil {
		return nil, err
	}

	return WrapConnection(conn), nil
}
