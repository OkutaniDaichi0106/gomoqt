package quic

import (
	"crypto/tls"

	"github.com/OkutaniDaichi0106/gomoqt/quic/internal"
	"github.com/OkutaniDaichi0106/gomoqt/quic/quicgo"
)

func ListenDefault(addr string, tlsConfig *tls.Config, quicConfig *Config) (EarlyListener, error) {
	return quicgo.ListenAddrEarly(addr, tlsConfig, quicConfig)
}

type EarlyListener = internal.EarlyListener
