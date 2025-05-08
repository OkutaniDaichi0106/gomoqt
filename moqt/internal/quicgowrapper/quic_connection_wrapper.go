package quicgowrapper

import (
	"context"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo "github.com/quic-go/quic-go"
)

func WrapConnection(conn quicgo.Connection) quic.Connection {
	return &rawQuicConnection{
		conn: conn,
	}
}

func UnWrapConnection(conn quic.Connection) quicgo.Connection {
	if l, ok := conn.(*rawQuicConnection); ok {
		return l.conn
	} else {
		return &quicgoConnection{conn: conn}
	}
}

var _ quic.Connection = (*rawQuicConnection)(nil)

type rawQuicConnection struct {
	conn quicgo.Connection
}

func (wrapper *rawQuicConnection) AcceptStream(ctx context.Context) (quic.Stream, error) {
	stream, err := wrapper.conn.AcceptStream(ctx)
	return &rawQuicStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) AcceptUniStream(ctx context.Context) (quic.ReceiveStream, error) {
	stream, err := wrapper.conn.AcceptUniStream(ctx)
	return &rawQuicReceiveStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) CloseWithError(code quic.ConnectionErrorCode, msg string) error {
	return wrapper.conn.CloseWithError(quicgo.ApplicationErrorCode(code), msg)
}
func (wrapper *rawQuicConnection) ConnectionState() quicgo.ConnectionState {
	return wrapper.conn.ConnectionState()
}
func (wrapper *rawQuicConnection) Context() context.Context {
	return wrapper.conn.Context()
}
func (wrapper *rawQuicConnection) LocalAddr() net.Addr {
	return wrapper.conn.LocalAddr()
}
func (wrapper *rawQuicConnection) OpenStream() (quic.Stream, error) {
	stream, err := wrapper.conn.OpenStream()
	return &rawQuicStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) OpenStreamSync(ctx context.Context) (quic.Stream, error) {
	stream, err := wrapper.conn.OpenStreamSync(ctx)
	return &rawQuicStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) OpenUniStream() (quic.SendStream, error) {
	stream, err := wrapper.conn.OpenUniStream()
	return &rawQuicSendStream{stream: stream}, err
}
func (wrapper *rawQuicConnection) OpenUniStreamSync(ctx context.Context) (quic.SendStream, error) {
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
