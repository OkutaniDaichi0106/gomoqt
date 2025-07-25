package webtransport

import (
	"context"
	"net"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type Server interface {
	Upgrade(w http.ResponseWriter, r *http.Request) (quic.Connection, error)
	ServeQUICConn(conn quic.Connection) error
	Serve(conn net.PacketConn) error
	Close() error
	Shutdown(context.Context) error
}
