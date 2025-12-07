package webtransportgo

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/okdaichi/gomoqt/quic"
	"github.com/okdaichi/gomoqt/webtransport"
	quicgo_webtransportgo "github.com/quic-go/webtransport-go"
)

var _ webtransport.DialAddrFunc = Dial

func Dial(ctx context.Context, addr string, header http.Header, tlsConfig *tls.Config) (*http.Response, quic.Connection, error) {
	d := quicgo_webtransportgo.Dialer{
		TLSClientConfig: tlsConfig,
	}
	rsp, wtsess, err := d.Dial(ctx, addr, header)

	return rsp, wrapSession(wtsess), err
}
