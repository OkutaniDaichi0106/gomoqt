package webtransport

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/okdaichi/gomoqt/quic"
)

// DialAddrFunc is a function type for establishing a WebTransport connection.
// It returns the HTTP response, the underlying QUIC connection, and any error.
type DialAddrFunc func(ctx context.Context, addr string, header http.Header, tlsConfig *tls.Config) (*http.Response, quic.Connection, error)
