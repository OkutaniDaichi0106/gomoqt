package webtransportgo

import (
	"context"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
	quicgo_webtransportgo "github.com/OkutaniDaichi0106/webtransport-go"
)

func Dial(ctx context.Context, addr string, header http.Header) (*http.Response, quic.Connection, error) {
	var d quicgo_webtransportgo.Dialer
	rsp, wtsess, err := d.Dial(ctx, addr, header)

	return rsp, wrapSession(wtsess), wrapError(err)
}
