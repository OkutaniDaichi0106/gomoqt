package webtransport

import (
	"context"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

// DialAddrFunc is a function type for establishing a WebTransport connection.
// It returns the HTTP response, the underlying QUIC connection, and any error.
type DialAddrFunc func(ctx context.Context, addr string, header http.Header) (*http.Response, quic.Connection, error)
