package quicgo

import (
	"context"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
	quicgo_quicgo "github.com/quic-go/quic-go"
)

func wrapConnection(conn *quicgo_quicgo.Conn) quic.Connection {
	if conn == nil {
		return nil
	}
	return &connWrapper{
		conn: conn,
	}
}

var _ quic.Connection = (*connWrapper)(nil)

type connWrapper struct {
	conn *quicgo_quicgo.Conn
}

func (wrapper *connWrapper) AcceptStream(ctx context.Context) (quic.Stream, error) {
	stream, err := wrapper.conn.AcceptStream(ctx)
	return &rawQuicStream{stream: stream}, wrapError(err)
}

func (wrapper *connWrapper) AcceptUniStream(ctx context.Context) (quic.ReceiveStream, error) {
	stream, err := wrapper.conn.AcceptUniStream(ctx)
	return &rawQuicReceiveStream{stream: stream}, wrapError(err)
}

func (wrapper *connWrapper) CloseWithError(code quic.ConnectionErrorCode, msg string) error {
	err := wrapper.conn.CloseWithError(quicgo_quicgo.ApplicationErrorCode(code), msg)
	return wrapError(err)
}

func (wrapper *connWrapper) ConnectionState() quic.ConnectionState {
	state := wrapper.conn.ConnectionState()
	return quic.ConnectionState{
		TLS:               state.TLS,
		SupportsDatagrams: state.SupportsDatagrams,
		Used0RTT:          state.Used0RTT,
		Version:           quic.Version(state.Version),
		GSO:               state.GSO,
	}
}

func (wrapper *connWrapper) Context() context.Context {
	return wrapper.conn.Context()
}

func (wrapper *connWrapper) LocalAddr() net.Addr {
	return wrapper.conn.LocalAddr()
}

func (wrapper *connWrapper) OpenStream() (quic.Stream, error) {
	stream, err := wrapper.conn.OpenStream()
	return &rawQuicStream{stream: stream}, wrapError(err)
}

func (wrapper *connWrapper) OpenStreamSync(ctx context.Context) (quic.Stream, error) {
	stream, err := wrapper.conn.OpenStreamSync(ctx)
	return &rawQuicStream{stream: stream}, wrapError(err)
}

func (wrapper *connWrapper) OpenUniStream() (quic.SendStream, error) {
	stream, err := wrapper.conn.OpenUniStream()
	return &rawQuicSendStream{stream: stream}, wrapError(err)
}

func (wrapper *connWrapper) OpenUniStreamSync(ctx context.Context) (quic.SendStream, error) {
	stream, err := wrapper.conn.OpenUniStreamSync(ctx)
	return &rawQuicSendStream{stream: stream}, wrapError(err)
}

// func (wrapper *connWrapper) ReceiveDatagram(ctx context.Context) ([]byte, error) {
// 	bytes, err := wrapper.conn.ReceiveDatagram(ctx)
// 	return bytes, wrapError(err)
// }

func (wrapper *connWrapper) RemoteAddr() net.Addr {
	return wrapper.conn.RemoteAddr()
}

// func (wrapper *connWrapper) SendDatagram(b []byte) error {
// 	return wrapError(wrapper.conn.SendDatagram(b))
// }

func (wrapper connWrapper) Unwrap() *quicgo_quicgo.Conn {
	return wrapper.conn
}
