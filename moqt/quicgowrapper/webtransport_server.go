package quicgowrapper

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/webtransport"
	quicgo "github.com/quic-go/quic-go"
	webtransportgo "github.com/quic-go/webtransport-go"
)

func WrapWebTransportServer(server *webtransportgo.Server) webtransport.Server {
	return &webTransportServer{
		server: server,
	}
}

var _ webtransport.Server = (*webTransportServer)(nil)

// webTransportServer is a wrapper for webtransport.Server
type webTransportServer struct {
	server *webtransportgo.Server
}

func (w *webTransportServer) TLSConfig() *tls.Config {
	return w.server.H3.TLSConfig
}

func (w *webTransportServer) QUICConfig() *quicgo.Config {
	return w.server.H3.QUICConfig
}

func (w *webTransportServer) SetTLSConfig(tlsConfig *tls.Config) {
	w.server.H3.TLSConfig = tlsConfig
}

func (wrapper *webTransportServer) SetQUICConfig(quicConfig *quicgo.Config) {
	wrapper.server.H3.QUICConfig = quicConfig
}

func (wrapper *webTransportServer) Upgrade(w http.ResponseWriter, r *http.Request) (quic.Connection, error) {
	wtsess, err := wrapper.server.Upgrade(w, r)
	if err != nil {
		return nil, err
	}

	return WrapWebTransportConnection(wtsess), nil
}

func (w *webTransportServer) ServeQUICConn(conn quic.Connection) error {
	return w.server.ServeQUICConn(UnWrapConnection(conn))
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
