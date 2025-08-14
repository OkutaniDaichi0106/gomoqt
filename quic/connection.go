package quic

import (
	"context"
	"net"

	"github.com/quic-go/quic-go"
)

type Connection interface {
	AcceptStream(ctx context.Context) (Stream, error)
	AcceptUniStream(ctx context.Context) (ReceiveStream, error)
	CloseWithError(code ApplicationErrorCode, msg string) error
	ConnectionState() ConnectionState
	Context() context.Context
	LocalAddr() net.Addr
	OpenStream() (Stream, error)
	OpenStreamSync(ctx context.Context) (Stream, error)
	OpenUniStream() (SendStream, error)
	OpenUniStreamSync(ctx context.Context) (str SendStream, err error)
	RemoteAddr() net.Addr
}

type ConnectionState = quic.ConnectionState

// struct {
// 	// TLS contains information about the TLS connection state, incl. the tls.ConnectionState.
// 	TLS tls.ConnectionState
// 	// SupportsDatagrams says if support for QUIC datagrams (RFC 9221) was negotiated.
// 	// This requires both nodes to support and enable the datagram extensions (via Config.EnableDatagrams).
// 	// If datagram support was negotiated, datagrams can be sent and received using the
// 	// SendDatagram and ReceiveDatagram methods on the Connection.
// 	SupportsDatagrams bool
// 	// Used0RTT says if 0-RTT resumption was used.
// 	Used0RTT bool
// 	// Version is the QUIC version of the QUIC connection.
// 	Version Version
// 	// GSO says if generic segmentation offload is used
// 	GSO bool
// }
