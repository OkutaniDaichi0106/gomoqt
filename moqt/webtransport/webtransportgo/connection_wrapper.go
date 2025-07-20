package webtransportgo

import (
	"context"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo_webtransportgo "github.com/OkutaniDaichi0106/webtransport-go"
)

type sessionWrapper struct {
	sess *quicgo_webtransportgo.Session
}

func WrapSession(wtconn *quicgo_webtransportgo.Session) quic.Connection {
	return &sessionWrapper{
		sess: wtconn,
	}
}

func UnWrapWebTransportConnection(conn quic.Connection) *quicgo_webtransportgo.Session {
	if wconn, ok := conn.(*sessionWrapper); ok {
		return wconn.sess
	}
	return nil
}

func (conn *sessionWrapper) AcceptStream(ctx context.Context) (quic.Stream, error) {
	stream, err := conn.sess.AcceptStream(ctx)
	return &streamWrapper{stream: stream}, err
}

func (conn *sessionWrapper) AcceptUniStream(ctx context.Context) (quic.ReceiveStream, error) {
	stream, err := conn.sess.AcceptUniStream(ctx)
	return &receiveStreamWrapper{stream: stream}, err
}

func (conn *sessionWrapper) CloseWithError(code quic.ConnectionErrorCode, msg string) error {
	return conn.sess.CloseWithError(quicgo_webtransportgo.SessionErrorCode(code), msg)
}

func (wrapper *sessionWrapper) ConnectionState() quic.ConnectionState {
	state := wrapper.sess.ConnectionState()
	return quic.ConnectionState{
		TLS:               state.TLS,
		SupportsDatagrams: state.SupportsDatagrams,
		Used0RTT:          state.Used0RTT,
		Version:           quic.Version(state.Version),
		GSO:               state.GSO,
	}
}

func (conn *sessionWrapper) Context() context.Context {
	return conn.sess.Context()
}

func (conn *sessionWrapper) LocalAddr() net.Addr {
	return conn.sess.LocalAddr()
}

func (conn *sessionWrapper) OpenStream() (quic.Stream, error) {
	stream, err := conn.sess.OpenStream()
	return &streamWrapper{stream: stream}, err
}

func (conn *sessionWrapper) OpenStreamSync(ctx context.Context) (quic.Stream, error) {
	stream, err := conn.sess.OpenStreamSync(ctx)
	return &streamWrapper{stream: stream}, err
}

func (conn *sessionWrapper) OpenUniStream() (quic.SendStream, error) {
	stream, err := conn.sess.OpenUniStream()
	return &sendStreamWrapper{stream: stream}, err
}

func (conn *sessionWrapper) OpenUniStreamSync(ctx context.Context) (quic.SendStream, error) {
	stream, err := conn.sess.OpenUniStreamSync(ctx)
	return &sendStreamWrapper{stream: stream}, err
}

func (conn *sessionWrapper) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return conn.sess.ReceiveDatagram(ctx)
}

func (conn *sessionWrapper) RemoteAddr() net.Addr {
	return conn.sess.RemoteAddr()
}

func (conn *sessionWrapper) SendDatagram(b []byte) error {
	return conn.sess.SendDatagram(b)
}
