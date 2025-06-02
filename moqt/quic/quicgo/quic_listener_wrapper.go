package quicgo

import (
	"crypto/tls"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo "github.com/quic-go/quic-go"
)

func Listen(addr string, tlsConfig *tls.Config, quicConfig *quic.Config) (quic.EarlyListener, error) {
	ln, err := quicgo.ListenAddrEarly(addr, tlsConfig, (*quicgo.Config)(quicConfig))
	return WrapListener(ln), err
}
