package moqt

import (
	"crypto/tls"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo "github.com/quic-go/quic-go"
)

var ListenQUICFunc func(addr string, tlsConf *tls.Config, config *quicgo.Config) (quic.EarlyListener, error)
