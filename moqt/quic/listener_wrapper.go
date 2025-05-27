package quic

import (
	"context"
	"net"

	quicgo "github.com/quic-go/quic-go"
)

var _ EarlyListener = (*listener)(nil)

func WrapListener(quicListener *quicgo.EarlyListener) EarlyListener {
	return &listener{
		quicListener: quicListener,
	}
}

func UnWrapListener(ln EarlyListener) *quicgo.EarlyListener {
	if l, ok := ln.(*listener); ok {
		return l.quicListener
	}
	return nil
}

type listener struct {
	quicListener *quicgo.EarlyListener
}

func (l *listener) Accept(ctx context.Context) (Connection, error) {
	conn, err := l.quicListener.Accept(ctx)
	if err != nil {
		return nil, err
	}

	return WrapConnection(conn), nil
}

func (l *listener) Addr() net.Addr {
	return l.quicListener.Addr()
}

func (l *listener) Close() error {
	return l.quicListener.Close()
}
