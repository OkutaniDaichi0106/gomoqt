package webtransport

import (
	"context"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

type DialAddrFunc func(ctx context.Context, addr string, header http.Header) (*http.Response, quic.Connection, error)
