package webtransportgo

import (
	"context"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/quic-go/webtransport-go"
)

type webtransportConnection struct {
	conn *webtransport.Session
}

func WrapWebTransportConnection(wtconn *webtransport.Session) quic.Connection {
	return &webtransportConnection{
		conn: wtconn,
	}
}

func UnWrapWebTransportConnection(conn quic.Connection) *webtransport.Session {
	if wconn, ok := conn.(*webtransportConnection); ok {
		return wconn.conn
	}
	return nil
}

func (conn *webtransportConnection) AcceptStream(ctx context.Context) (quic.Stream, error) {
	stream, err := conn.conn.AcceptStream(ctx)
	return &webtransportStream{stream: stream}, err
}

func (conn *webtransportConnection) AcceptUniStream(ctx context.Context) (quic.ReceiveStream, error) {
	stream, err := conn.conn.AcceptUniStream(ctx)
	return &webtransportReceiveStream{stream: stream}, err
}

func (conn *webtransportConnection) CloseWithError(code quic.ConnectionErrorCode, msg string) error {
	return conn.conn.CloseWithError(webtransport.SessionErrorCode(code), msg)
}

func (conn *webtransportConnection) ConnectionState() quic.ConnectionState {
	return quic.ConnectionState(conn.conn.ConnectionState())
}

func (conn *webtransportConnection) Context() context.Context {
	return conn.conn.Context()
}

func (conn *webtransportConnection) LocalAddr() net.Addr {
	return conn.conn.LocalAddr()
}

func (conn *webtransportConnection) OpenStream() (quic.Stream, error) {
	stream, err := conn.conn.OpenStream()
	return &webtransportStream{stream: stream}, err
}

func (conn *webtransportConnection) OpenStreamSync(ctx context.Context) (quic.Stream, error) {
	stream, err := conn.conn.OpenStreamSync(ctx)
	return &webtransportStream{stream: stream}, err
}

func (conn *webtransportConnection) OpenUniStream() (quic.SendStream, error) {
	stream, err := conn.conn.OpenUniStream()
	return &webtransportSendStream{stream: stream}, err
}

func (conn *webtransportConnection) OpenUniStreamSync(ctx context.Context) (quic.SendStream, error) {
	stream, err := conn.conn.OpenUniStreamSync(ctx)
	return &webtransportSendStream{stream: stream}, err
}

func (conn *webtransportConnection) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return conn.conn.ReceiveDatagram(ctx)
}

func (conn *webtransportConnection) RemoteAddr() net.Addr {
	return conn.conn.RemoteAddr()
}

func (conn *webtransportConnection) SendDatagram(b []byte) error {
	return conn.conn.SendDatagram(b)
}
