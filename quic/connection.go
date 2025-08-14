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
