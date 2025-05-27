package webtransport

import (
	"context"
	"net"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	webtransportgo "github.com/quic-go/webtransport-go"
)

func WrapWebTransportServer(server *webtransportgo.Server) Server {
	return &webTransportServer{
		server: server,
	}
}

var _ Server = (*webTransportServer)(nil)

// webTransportServer is a wrapper for Server
type webTransportServer struct {
	server *webtransportgo.Server
}

func (wrapper *webTransportServer) Upgrade(w http.ResponseWriter, r *http.Request) (quic.Connection, error) {
	wtsess, err := wrapper.server.Upgrade(w, r)
	if err != nil {
		return nil, err
	}

	return WrapWebTransportConnection(wtsess), nil
}

func (w *webTransportServer) ServeQUICConn(conn quic.Connection) error {
	return w.server.ServeQUICConn(quic.UnWrapConnection(conn))
}

func (w *webTransportServer) Serve(conn net.PacketConn) error {

	return w.server.Serve(conn)
}

func (w *webTransportServer) Close() error {
	return w.server.Close()
}

func (w *webTransportServer) Shutdown(ctx context.Context) error {
	// TODO: Implement Shutdown logic if needed
	return nil
}
