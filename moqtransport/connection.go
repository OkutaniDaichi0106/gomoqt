package moqtransport

import (
	"context"
	"net"
	"net/url"

	"github.com/quic-go/quic-go"
)

type Connection interface {
	AcceptStream(ctx context.Context) (Stream, error)
	AcceptUniStream(ctx context.Context) (ReceiveStream, error)
	CloseWithError(code SessionErrorCode, msg string) error
	ConnectionState() quic.ConnectionState
	Context() context.Context
	LocalAddr() net.Addr
	OpenStream() (Stream, error)
	OpenStreamSync(ctx context.Context) (Stream, error)
	OpenUniStream() (SendStream, error)
	OpenUniStreamSync(ctx context.Context) (str SendStream, err error)
	ReceiveDatagram(ctx context.Context) ([]byte, error)
	RemoteAddr() net.Addr
	SendDatagram(b []byte) error
	URL() url.URL
}
