package webtransport

import (
	"context"
	"net"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type Server interface {
	Upgrade(w http.ResponseWriter, r *http.Request) (quic.Connection, error)
	Handle(path string, handler Handler)
	HandleFunc(path string, f HandlerFunc)
	ServeQUICConn(conn quic.Connection) error
	Serve(conn net.PacketConn) error
	Close() error
	Shutdown(context.Context) error
}

type Handler interface {
	Handle(r *http.Request, conn quic.Connection)
}

type HandlerFunc func(r *http.Request, conn quic.Connection)

func (f HandlerFunc) Handle(r *http.Request, conn quic.Connection) {
	f(r, conn)
}
