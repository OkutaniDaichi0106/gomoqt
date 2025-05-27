package webtransport

import (
	"context"
	"net"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/quic-go/quic-go/http3"
	webtransportgo "github.com/quic-go/webtransport-go"
)

func NewDefaultServer(addr string) Server {
	wtserver := &webtransportgo.Server{
		H3: http3.Server{
			Addr: addr,
		},
	}

	return WrapWebTransportServer(wtserver)
}

type Server interface {
	Upgrade(w http.ResponseWriter, r *http.Request) (quic.Connection, error)
	ServeQUICConn(conn quic.Connection) error
	Serve(conn net.PacketConn) error
	Close() error
	Shutdown(context.Context) error
}
