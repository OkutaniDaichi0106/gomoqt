package quic

import (
	"crypto/tls"

	quicgo "github.com/quic-go/quic-go"
)

var ListenQUICFunc func(addr string, tlsConf *tls.Config, config *Config) (EarlyListener, error) = defaultListenQUICFunc

var defaultListenQUICFunc = func(addr string, tlsConfig *tls.Config, quicConfig *Config) (EarlyListener, error) {
	ln, err := quicgo.ListenAddrEarly(addr, tlsConfig, (*quicgo.Config)(quicConfig))
	return WrapListener(ln), err
}
