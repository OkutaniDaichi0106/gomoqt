package webtransport

import (
	"context"
	"net/http"

	"github.com/okdaichi/gomoqt/quic"
)

// Server handles WebTransport connections over HTTP/3.
type Server interface {
	// Upgrade upgrades an HTTP request to a WebTransport connection.
	Upgrade(w http.ResponseWriter, r *http.Request) (quic.Connection, error)

	// ServeQUICConn serves a QUIC connection as a WebTransport session.
	ServeQUICConn(conn quic.Connection) error

	// Close immediately closes the server and all active connections.
	Close() error

	// Shutdown gracefully shuts down the server without interrupting active connections.
	Shutdown(context.Context) error
}
