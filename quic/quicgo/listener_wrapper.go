package quicgo

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
	quicgo_quicgo "github.com/quic-go/quic-go"
)

func ListenAddrEarly(addr string, tlsConfig *tls.Config, quicConfig *quic.Config) (quic.EarlyListener, error) {
	ln, err := quicgo_quicgo.ListenAddrEarly(addr, tlsConfig, quicConfig)
	return wrapListener(ln), err
}

var _ quic.EarlyListener = (*listenerWrapper)(nil)

func wrapListener(quicListener *quicgo_quicgo.EarlyListener) quic.EarlyListener {
	return &listenerWrapper{
		listener: quicListener,
	}
}

type listenerWrapper struct {
	listener *quicgo_quicgo.EarlyListener
}

func (wrapper *listenerWrapper) Accept(ctx context.Context) (quic.Connection, error) {
	conn, err := wrapper.listener.Accept(ctx)
	return wrapConnection(conn), err
}

func (wrapper *listenerWrapper) Addr() net.Addr {
	return wrapper.listener.Addr()
}

func (wrapper *listenerWrapper) Close() error {
	return wrapper.listener.Close()
}
