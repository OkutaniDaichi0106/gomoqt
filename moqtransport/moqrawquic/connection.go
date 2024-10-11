package moqrawquic

import (
	"context"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/quic-go/quic-go"
)

func NewMOQTConnection(conn quic.Connection) moqtransport.Connection {
	return &rawQuicConnection{
		conn: conn,
	}
}

type rawQuicConnection struct {
	conn quic.Connection
}

func (wrapper *rawQuicConnection) AcceptStream(ctx context.Context) (moqtransport.Stream, error) {
	stream, err := wrapper.conn.AcceptStream(ctx)
	return &rawQuicStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) AcceptUniStream(ctx context.Context) (moqtransport.ReceiveStream, error) {
	stream, err := wrapper.conn.AcceptUniStream(ctx)
	return &rawQuicReceiveStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) CloseWithError(code moqtransport.SessionErrorCode, msg string) error {
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
func (wrapper *rawQuicConnection) OpenStream() (moqtransport.Stream, error) {
	stream, err := wrapper.conn.OpenStream()
	return &rawQuicStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) OpenStreamSync(ctx context.Context) (moqtransport.Stream, error) {
	stream, err := wrapper.conn.OpenStreamSync(ctx)
	return &rawQuicStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) OpenUniStream() (moqtransport.SendStream, error) {
	stream, err := wrapper.conn.OpenUniStream()
	return &rawQuicSendStreamWrapper{stream: stream}, err
}
func (wrapper *rawQuicConnection) OpenUniStreamSync(ctx context.Context) (moqtransport.SendStream, error) {
	stream, err := wrapper.conn.OpenUniStreamSync(ctx)
	return &rawQuicSendStreamWrapper{stream: stream}, err
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
