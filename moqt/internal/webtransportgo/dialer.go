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
// connection under it; nil is the library's defaults.
func Dial(ctx context.Context, addr string, header http.Header, tlsConfig *tls.Config, quicConfig *quic.Config, appProtocols []string) (*http.Response, transport.WebTransportSession, error) {
	dialer := quicgo_webtransportgo.Dialer{
		TLSClientConfig:      tlsConfig,
		QUICConfig:           dialQUICConfig(quicConfig),
		ApplicationProtocols: appProtocols,
	}
	rsp, wtsess, err := dialer.Dial(ctx, addr, header)

	return rsp, wrapSession(wtsess), err
}

// dialQUICConfig returns the QUIC configuration a WebTransport dial uses for
// c: nil for nil, and otherwise a copy with datagrams and stream reset with
// partial delivery on. WebTransport needs both, and the library refuses a
// configuration without them, so a caller's own configuration need not set
// them, as a server's need not.
func dialQUICConfig(c *quic.Config) *quic.Config {
	if c == nil {
		return nil
	}
	c = c.Clone()
	c.EnableDatagrams = true
	c.EnableStreamResetPartialDelivery = true
	return c
}
