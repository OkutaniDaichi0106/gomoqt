package moqwebtransport

import (
	"context"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

type webtransportConnection struct {
	conn *webtransport.Session
}

func NewMOQTConnection(conn *webtransport.Session) moqtransport.Connection {
	return &webtransportConnection{
		conn: conn,
	}
}

func (conn *webtransportConnection) AcceptStream(ctx context.Context) (moqtransport.Stream, error) {
	stream, err := conn.conn.AcceptStream(ctx)
	return &webtransportStream{str: stream}, err
}
func (conn *webtransportConnection) AcceptUniStream(ctx context.Context) (moqtransport.ReceiveStream, error) {
	stream, err := conn.conn.AcceptUniStream(ctx)
	return &webtransportReceiveStream{innerReceiveStream: stream}, err
}
func (conn *webtransportConnection) CloseWithError(code moqtransport.SessionErrorCode, msg string) error {
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
func (conn *webtransportConnection) OpenStream() (moqtransport.Stream, error) {
	stream, err := conn.conn.OpenStream()
	return &webtransportStream{str: stream}, err
}
func (conn *webtransportConnection) OpenStreamSync(ctx context.Context) (moqtransport.Stream, error) {
	stream, err := conn.conn.OpenStreamSync(ctx)
	return &webtransportStream{str: stream}, err
}
func (conn *webtransportConnection) OpenUniStream() (moqtransport.SendStream, error) {
	stream, err := conn.conn.OpenUniStream()
	return &webtransportSendStream{innerSendStream: stream}, err
}
func (conn *webtransportConnection) OpenUniStreamSync(ctx context.Context) (moqtransport.SendStream, error) {
	stream, err := conn.conn.OpenUniStreamSync(ctx)
	return &webtransportSendStream{innerSendStream: stream}, err
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
