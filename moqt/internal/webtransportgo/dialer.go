package webtransportgo

import (
	"context"
	"crypto/tls"
	"net/http"

	quicgo_webtransportgo "github.com/okdaichi/webtransport-go"
	"github.com/quic-go/quic-go"
	"github.com/qumo-dev/gomoqt/transport"
)

// Dial opens a WebTransport session to addr. quicConfig configures the QUIC
// connection under it and is passed to the library as it is: nil is the
// library's defaults, and any other must have datagrams and stream reset with
// partial delivery on, or the library refuses it.
func Dial(ctx context.Context, addr string, header http.Header, tlsConfig *tls.Config, quicConfig *quic.Config, appProtocols []string) (*http.Response, transport.WebTransportSession, error) {
	dialer := quicgo_webtransportgo.Dialer{
		TLSClientConfig:      tlsConfig,
		QUICConfig:           quicConfig,
		ApplicationProtocols: appProtocols,
	}
	rsp, wtsess, err := dialer.Dial(ctx, addr, header)

	return rsp, wrapSession(wtsess), err
}
