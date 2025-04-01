package webtransport

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo "github.com/quic-go/quic-go"
)

type Server interface {
	TLSConfig() *tls.Config
	QUICConfig() *quicgo.Config

	SetTLSConfig(tlsConfig *tls.Config)
	SetQUICConfig(quicConfig *quicgo.Config)

	Upgrade(w http.ResponseWriter, r *http.Request) (quic.Connection, error)
	ServeQUICConn(conn quic.Connection) error
	Serve(conn net.PacketConn) error
	Close() error
	Shutdown(context.Context) error
}
