package moqt

import (
	"context"
	"net"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

type webtransportConnection struct {
	conn *webtransport.Session
}

func NewMOWTConnection(wtconn *webtransport.Session) Connection {
	return &webtransportConnection{
		conn: wtconn,
	}
}

func (conn *webtransportConnection) AcceptStream(ctx context.Context) (Stream, error) {
	stream, err := conn.conn.AcceptStream(ctx)
	return &webtransportStream{stream: stream}, err
}
func (conn *webtransportConnection) AcceptUniStream(ctx context.Context) (ReceiveStream, error) {
	stream, err := conn.conn.AcceptUniStream(ctx)
	return &webtransportReceiveStream{stream: stream}, err
}
func (conn *webtransportConnection) CloseWithError(code SessionErrorCode, msg string) error {
	return conn.conn.CloseWithError(webtransport.SessionErrorCode(code), msg)
}
func (conn *webtransportConnection) ConnectionState() quic.ConnectionState {
	return conn.conn.ConnectionState()
}
func (conn *webtransportConnection) Context() context.Context {
	return conn.conn.Context()
}
func (conn *webtransportConnection) LocalAddr() net.Addr {
	return conn.conn.LocalAddr()
}
func (conn *webtransportConnection) OpenStream() (Stream, error) {
	stream, err := conn.conn.OpenStream()
	return &webtransportStream{stream: stream}, err
}
func (conn *webtransportConnection) OpenStreamSync(ctx context.Context) (Stream, error) {
	stream, err := conn.conn.OpenStreamSync(ctx)
	return &webtransportStream{stream: stream}, err
}
func (conn *webtransportConnection) OpenUniStream() (SendStream, error) {
	stream, err := conn.conn.OpenUniStream()
	return &webtransportSendStream{stream: stream}, err
}
func (conn *webtransportConnection) OpenUniStreamSync(ctx context.Context) (SendStream, error) {
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
