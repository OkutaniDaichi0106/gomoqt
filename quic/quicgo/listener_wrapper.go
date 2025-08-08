package quicgo

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/quic/internal"
	quicgo_quicgo "github.com/quic-go/quic-go"
)

func ListenAddrEarly(addr string, tlsConfig *tls.Config, quicConfig *internal.Config) (internal.EarlyListener, error) {
	ln, err := quicgo_quicgo.ListenAddrEarly(addr, tlsConfig, quicConfig)
	return wrapListener(ln), wrapError(err)
}

var _ internal.EarlyListener = (*listenerWrapper)(nil)

func wrapListener(quicListener *quicgo_quicgo.EarlyListener) internal.EarlyListener {
	return &listenerWrapper{
		listener: quicListener,
	}
}

type listenerWrapper struct {
	listener *quicgo_quicgo.EarlyListener
}

func (wrapper *listenerWrapper) Accept(ctx context.Context) (internal.Connection, error) {
	conn, err := wrapper.listener.Accept(ctx)
	return wrapConnection(conn), wrapError(err)
}

func (wrapper *listenerWrapper) Addr() net.Addr {
	return wrapper.listener.Addr()
}

func (wrapper *listenerWrapper) Close() error {
	return wrapError(wrapper.listener.Close())
}
