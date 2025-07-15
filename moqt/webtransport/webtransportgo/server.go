package webtransportgo

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/webtransport"
	quicgo_quicgo "github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	quicgo_webtransportgo "github.com/quic-go/webtransport-go"
)

func NewDefaultServer(addr string, tlsConfig *tls.Config, quicConfig *quic.Config, checkOrigin func(r *http.Request) bool) webtransport.Server {
	wtserver := &quicgo_webtransportgo.Server{
		H3: http3.Server{
			Addr:       addr,
			TLSConfig:  tlsConfig,
			QUICConfig: quicConfig,
		},
		CheckOrigin: checkOrigin,
	}

	return WrapServer(wtserver)
}

func WrapServer(server *quicgo_webtransportgo.Server) webtransport.Server {
	return &serverWrapper{
		server: server,
	}
}

var _ webtransport.Server = (*serverWrapper)(nil)

// serverWrapper is a wrapper for Server
type serverWrapper struct {
	server *quicgo_webtransportgo.Server
}

func (wrapper *serverWrapper) Upgrade(w http.ResponseWriter, r *http.Request) (quic.Connection, error) {
	wtsess, err := wrapper.server.Upgrade(w, r)
	if err != nil {
		return nil, err
	}

	return WrapSession(wtsess), nil
}

func (w *serverWrapper) ServeQUICConn(conn quic.Connection) error {
	if conn == nil {
		return nil
	}
	if unwrapper, ok := conn.(interface {
		Unwrap() *quicgo_quicgo.Conn
	}); ok {
		return w.server.ServeQUICConn(unwrapper.Unwrap())
	}
	return errors.New("invalid connection type: expected a wrapped quic-go connection with Unwrap() method")
}

func (w *serverWrapper) Serve(conn net.PacketConn) error {

	return w.server.Serve(conn)
}

func (w *serverWrapper) Close() error {
	return w.server.Close()
}

func (w *serverWrapper) Shutdown(ctx context.Context) error {
	// Implement a proper shutdown logic that passes the context to the server
	closeCh := make(chan struct{})

	// Close the server in a separate goroutine
	go func() {
		err := w.server.Close()
		if err != nil {
			// Log the error if needed
		}
		close(closeCh)
	}()

	// Wait for either the context to be done or the close to complete
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-closeCh:
		return nil
	}
}
