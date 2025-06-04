package quicgo

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo_quicgo "github.com/quic-go/quic-go"
)

func Listen(addr string, tlsConfig *tls.Config, quicConfig *quic.Config) (quic.EarlyListener, error) {
	ln, err := quicgo_quicgo.ListenAddrEarly(addr, tlsConfig, (*quicgo_quicgo.Config)(quicConfig))
	return WrapListener(ln), WrapError(err)
}

var _ quic.EarlyListener = (*listenerWrapper)(nil)

func WrapListener(quicListener *quicgo_quicgo.EarlyListener) quic.EarlyListener {
	return &listenerWrapper{
		quicListener: quicListener,
	}
}

type listenerWrapper struct {
	quicListener *quicgo_quicgo.EarlyListener
}

func (l *listenerWrapper) Accept(ctx context.Context) (quic.Connection, error) {
	conn, err := l.quicListener.Accept(ctx)
	return WrapConnection(conn), WrapError(err)
}

func (l *listenerWrapper) Addr() net.Addr {
	return l.quicListener.Addr()
}

func (l *listenerWrapper) Close() error {
	return WrapError(l.quicListener.Close())
}
