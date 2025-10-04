package webtransportgo

import (
	"context"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
	"github.com/OkutaniDaichi0106/gomoqt/webtransport"
	quicgo_webtransportgo "github.com/OkutaniDaichi0106/webtransport-go"
)

var _ webtransport.DialAddrFunc = Dial

func Dial(ctx context.Context, addr string, header http.Header) (*http.Response, quic.Connection, error) {
	var d quicgo_webtransportgo.Dialer
	rsp, wtsess, err := d.Dial(ctx, addr, header)

	return rsp, wrapSession(wtsess), err
}
