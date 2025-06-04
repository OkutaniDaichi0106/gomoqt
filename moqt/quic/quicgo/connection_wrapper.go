package quicgo

import (
	"context"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo_quicgo "github.com/quic-go/quic-go"
)

func WrapConnection(conn quicgo_quicgo.Connection) quic.Connection {
	if conn == nil {
		return nil
	}
	return &connectionWrapper{
		conn: conn,
	}
}

var _ quic.Connection = (*connectionWrapper)(nil)

type connectionWrapper struct {
	conn quicgo_quicgo.Connection
}

func (wrapper *connectionWrapper) AcceptStream(ctx context.Context) (quic.Stream, error) {
	stream, err := wrapper.conn.AcceptStream(ctx)
	return &rawQuicStream{stream: stream}, WrapError(err)
}
func (wrapper *connectionWrapper) AcceptUniStream(ctx context.Context) (quic.ReceiveStream, error) {
	stream, err := wrapper.conn.AcceptUniStream(ctx)
	return &rawQuicReceiveStream{stream: stream}, WrapError(err)
}
func (wrapper *connectionWrapper) CloseWithError(code quic.ConnectionErrorCode, msg string) error {
	return WrapError(wrapper.conn.CloseWithError(quicgo_quicgo.ApplicationErrorCode(code), msg))
}
func (wrapper *connectionWrapper) ConnectionState() quic.ConnectionState {
	return quic.ConnectionState(wrapper.conn.ConnectionState())
}
func (wrapper *connectionWrapper) Context() context.Context {
	return wrapper.conn.Context()
}
func (wrapper *connectionWrapper) LocalAddr() net.Addr {
	return wrapper.conn.LocalAddr()
}
func (wrapper *connectionWrapper) OpenStream() (quic.Stream, error) {
	stream, err := wrapper.conn.OpenStream()
	return &rawQuicStream{stream: stream}, WrapError(err)
}
func (wrapper *connectionWrapper) OpenStreamSync(ctx context.Context) (quic.Stream, error) {
	stream, err := wrapper.conn.OpenStreamSync(ctx)
	return &rawQuicStream{stream: stream}, WrapError(err)
}
func (wrapper *connectionWrapper) OpenUniStream() (quic.SendStream, error) {
	stream, err := wrapper.conn.OpenUniStream()
	return &rawQuicSendStream{stream: stream}, WrapError(err)
}
func (wrapper *connectionWrapper) OpenUniStreamSync(ctx context.Context) (quic.SendStream, error) {
	stream, err := wrapper.conn.OpenUniStreamSync(ctx)
	return &rawQuicSendStream{stream: stream}, WrapError(err)
}
func (wrapper *connectionWrapper) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	bytes, err := wrapper.conn.ReceiveDatagram(ctx)
	return bytes, WrapError(err)
}
func (wrapper *connectionWrapper) RemoteAddr() net.Addr {
	return wrapper.conn.RemoteAddr()
}
func (wrapper *connectionWrapper) SendDatagram(b []byte) error {
	return WrapError(wrapper.conn.SendDatagram(b))
}

func (wrapper connectionWrapper) Unwrap() quicgo_quicgo.Connection {
	return wrapper.conn
}
