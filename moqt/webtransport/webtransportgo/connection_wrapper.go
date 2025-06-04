package webtransportgo

import (
	"context"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo_webtransportgo "github.com/quic-go/webtransport-go"
)

type sessionWrapper struct {
	conn *quicgo_webtransportgo.Session
}

func WrapSession(wtconn *quicgo_webtransportgo.Session) quic.Connection {
	return &sessionWrapper{
		conn: wtconn,
	}
}

func UnWrapWebTransportConnection(conn quic.Connection) *quicgo_webtransportgo.Session {
	if wconn, ok := conn.(*sessionWrapper); ok {
		return wconn.conn
	}
	return nil
}

func (conn *sessionWrapper) AcceptStream(ctx context.Context) (quic.Stream, error) {
	stream, err := conn.conn.AcceptStream(ctx)
	return &streamWrapper{stream: stream}, err
}

func (conn *sessionWrapper) AcceptUniStream(ctx context.Context) (quic.ReceiveStream, error) {
	stream, err := conn.conn.AcceptUniStream(ctx)
	return &receiveStreamWrapper{stream: stream}, err
}

func (conn *sessionWrapper) CloseWithError(code quic.ConnectionErrorCode, msg string) error {
	return conn.conn.CloseWithError(quicgo_webtransportgo.SessionErrorCode(code), msg)
}

func (conn *sessionWrapper) ConnectionState() quic.ConnectionState {
	return quic.ConnectionState(conn.conn.ConnectionState())
}

func (conn *sessionWrapper) Context() context.Context {
	return conn.conn.Context()
}

func (conn *sessionWrapper) LocalAddr() net.Addr {
	return conn.conn.LocalAddr()
}

func (conn *sessionWrapper) OpenStream() (quic.Stream, error) {
	stream, err := conn.conn.OpenStream()
	return &streamWrapper{stream: stream}, err
}

func (conn *sessionWrapper) OpenStreamSync(ctx context.Context) (quic.Stream, error) {
	stream, err := conn.conn.OpenStreamSync(ctx)
	return &streamWrapper{stream: stream}, err
}

func (conn *sessionWrapper) OpenUniStream() (quic.SendStream, error) {
	stream, err := conn.conn.OpenUniStream()
	return &sendStreamWrapper{stream: stream}, err
}

func (conn *sessionWrapper) OpenUniStreamSync(ctx context.Context) (quic.SendStream, error) {
	stream, err := conn.conn.OpenUniStreamSync(ctx)
	return &sendStreamWrapper{stream: stream}, err
}

func (conn *sessionWrapper) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return conn.conn.ReceiveDatagram(ctx)
}

func (conn *sessionWrapper) RemoteAddr() net.Addr {
	return conn.conn.RemoteAddr()
}

func (conn *sessionWrapper) SendDatagram(b []byte) error {
	return conn.conn.SendDatagram(b)
}
