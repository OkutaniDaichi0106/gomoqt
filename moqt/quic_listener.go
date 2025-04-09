package moqt

import (
	"crypto/tls"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/quicgowrapper"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo "github.com/quic-go/quic-go"
)

var ListenQUICFunc func(addr string, tlsConf *tls.Config, config *quicgo.Config) (quic.EarlyListener, error) = defaultListenQUICFunc

var defaultListenQUICFunc = func(addr string, tlsConfig *tls.Config, quicConfig *quicgo.Config) (quic.EarlyListener, error) {
	ln, err := quicgo.ListenAddrEarly(addr, tlsConfig, quicConfig)
	return quicgowrapper.WrapListener(ln), err
}
