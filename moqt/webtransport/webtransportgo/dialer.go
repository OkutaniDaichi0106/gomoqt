package webtransportgo

import (
	"context"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic/quicgo"
	quicgo_webtransportgo "github.com/quic-go/webtransport-go"
)

func Dial(ctx context.Context, addr string, header http.Header) (*http.Response, quic.Connection, error) {
	var d quicgo_webtransportgo.Dialer
	rsp, wtsess, err := d.Dial(ctx, addr, header)

	return rsp, WrapSession(wtsess), quicgo.WrapError(err)
}
