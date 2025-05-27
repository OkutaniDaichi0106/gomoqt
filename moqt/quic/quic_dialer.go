package quic

import (
	"context"
	"crypto/tls"

	quicgo "github.com/quic-go/quic-go"
)

var DialFunc = defaultDialQUICFunc

var defaultDialQUICFunc = func(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *quicgo.Config) (Connection, error) {
	conn, err := quicgo.DialAddrEarly(ctx, addr, tlsConfig, quicConfig)
	if err != nil {
		return nil, err
	}

	return WrapConnection(conn), nil
}
