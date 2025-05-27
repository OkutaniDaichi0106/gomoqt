package quic

import (
	"crypto/tls"

	quicgo "github.com/quic-go/quic-go"
)

var ListenQUICFunc func(addr string, tlsConf *tls.Config, config *quicgo.Config) (EarlyListener, error) = defaultListenQUICFunc

var defaultListenQUICFunc = func(addr string, tlsConfig *tls.Config, quicConfig *quicgo.Config) (EarlyListener, error) {
	ln, err := quicgo.ListenAddrEarly(addr, tlsConfig, quicConfig)
	return WrapListener(ln), err
}
