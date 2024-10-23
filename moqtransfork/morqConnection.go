package moqtransfork

import (
	"context"
	"net"

	"github.com/quic-go/quic-go"
)

func NewMORQConnection(conn quic.Connection) Connection {
	return &rawQuicConnection{
		conn: conn,
	}
}

type rawQuicConnection struct {
	conn quic.Connection
}

func (wrapper *rawQuicConnection) AcceptStream(ctx context.Context) (Stream, error) {
	stream, err := wrapper.conn.AcceptStream(ctx)
	return &rawQuicStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) AcceptUniStream(ctx context.Context) (ReceiveStream, error) {
	stream, err := wrapper.conn.AcceptUniStream(ctx)
	return &rawQuicReceiveStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) CloseWithError(code SessionErrorCode, msg string) error {
	return wrapper.conn.CloseWithError(quic.ApplicationErrorCode(code), msg)
}
func (wrapper *rawQuicConnection) ConnectionState() quic.ConnectionState {
	return wrapper.conn.ConnectionState()
}
func (wrapper *rawQuicConnection) Context() context.Context {
	return wrapper.conn.Context()
}
func (wrapper *rawQuicConnection) LocalAddr() net.Addr {
	return wrapper.conn.LocalAddr()
}
func (wrapper *rawQuicConnection) OpenStream() (Stream, error) {
	stream, err := wrapper.conn.OpenStream()
	return &rawQuicStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) OpenStreamSync(ctx context.Context) (Stream, error) {
	stream, err := wrapper.conn.OpenStreamSync(ctx)
	return &rawQuicStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) OpenUniStream() (SendStream, error) {
	stream, err := wrapper.conn.OpenUniStream()
	return &rawQuicSendStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) OpenUniStreamSync(ctx context.Context) (SendStream, error) {
	stream, err := wrapper.conn.OpenUniStreamSync(ctx)
	return &rawQuicSendStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return wrapper.conn.ReceiveDatagram(ctx)
}
func (wrapper *rawQuicConnection) RemoteAddr() net.Addr {
	return wrapper.conn.RemoteAddr()
}
func (wrapper *rawQuicConnection) SendDatagram(b []byte) error {
	return wrapper.conn.SendDatagram(b)
}
