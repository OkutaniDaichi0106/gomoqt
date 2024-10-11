package moqtransport

import (
	"context"
	"net"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

/*
 * Transport Session: The wrapper of the both raw QUIC connection and the WebTransport session
 */
type TransportSession interface {
	AcceptStream(ctx context.Context) (Stream, error)
	AcceptUniStream(ctx context.Context) (ReceiveStream, error)
	CloseWithError(code SessionErrorCode, msg string) error
	ConnectionState() quic.ConnectionState
	Context() context.Context
	LocalAddr() net.Addr
	OpenStream() (Stream, error)
	OpenStreamSync(ctx context.Context) (Stream, error)
	OpenUniStream() (SendStream, error)
	OpenUniStreamSync(ctx context.Context) (str SendStream, err error)
	ReceiveDatagram(ctx context.Context) ([]byte, error)
	RemoteAddr() net.Addr
	SendDatagram(b []byte) error
}

type rawQuicConnectionWrapper struct {
	innerSession quic.Connection
}

func (sess *rawQuicConnectionWrapper) AcceptStream(ctx context.Context) (Stream, error) {
	stream, err := sess.innerSession.AcceptStream(ctx)
	return &rawQuicStreamWrapper{innerStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) AcceptUniStream(ctx context.Context) (ReceiveStream, error) {
	stream, err := sess.innerSession.AcceptUniStream(ctx)
	return &rawQuicReceiveStreamWrapper{innerReceiveStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) CloseWithError(code SessionErrorCode, msg string) error {
	return sess.innerSession.CloseWithError(quic.ApplicationErrorCode(code), msg)
}
func (sess *rawQuicConnectionWrapper) ConnectionState() quic.ConnectionState {
	return sess.innerSession.ConnectionState()
}
func (sess *rawQuicConnectionWrapper) Context() context.Context {
	return sess.innerSession.Context()
}
func (sess *rawQuicConnectionWrapper) LocalAddr() net.Addr {
	return sess.innerSession.LocalAddr()
}
func (sess *rawQuicConnectionWrapper) OpenStream() (Stream, error) {
	stream, err := sess.innerSession.OpenStream()
	return &rawQuicStreamWrapper{innerStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) OpenStreamSync(ctx context.Context) (Stream, error) {
	stream, err := sess.innerSession.OpenStreamSync(ctx)
	return &rawQuicStreamWrapper{innerStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) OpenUniStream() (SendStream, error) {
	stream, err := sess.innerSession.OpenUniStream()
	return &rawQuicSendStreamWrapper{innerSendStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) OpenUniStreamSync(ctx context.Context) (SendStream, error) {
	stream, err := sess.innerSession.OpenUniStreamSync(ctx)
	return &rawQuicSendStreamWrapper{innerSendStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return sess.innerSession.ReceiveDatagram(ctx)
}
func (sess *rawQuicConnectionWrapper) RemoteAddr() net.Addr {
	return sess.innerSession.RemoteAddr()
}
func (sess *rawQuicConnectionWrapper) SendDatagram(b []byte) error {
	return sess.innerSession.SendDatagram(b)
}

type webtransportSessionWrapper struct {
	innerSession *webtransport.Session
}

func (sess *webtransportSessionWrapper) AcceptStream(ctx context.Context) (Stream, error) {
	stream, err := sess.innerSession.AcceptStream(ctx)
	return &webtransportStreamWrapper{innerStream: stream}, err
}
func (sess *webtransportSessionWrapper) AcceptUniStream(ctx context.Context) (ReceiveStream, error) {
	stream, err := sess.innerSession.AcceptUniStream(ctx)
	return &webtransportReceiveStreamWrapper{innerReceiveStream: stream}, err
}
func (sess *webtransportSessionWrapper) CloseWithError(code SessionErrorCode, msg string) error {
	return sess.innerSession.CloseWithError(webtransport.SessionErrorCode(code), msg)
}
func (sess *webtransportSessionWrapper) ConnectionState() quic.ConnectionState {
	return sess.innerSession.ConnectionState()
}
func (sess *webtransportSessionWrapper) Context() context.Context {
	return sess.innerSession.Context()
}
func (sess *webtransportSessionWrapper) LocalAddr() net.Addr {
	return sess.innerSession.LocalAddr()
}
func (sess *webtransportSessionWrapper) OpenStream() (Stream, error) {
	stream, err := sess.innerSession.OpenStream()
	return &webtransportStreamWrapper{innerStream: stream}, err
}
func (sess *webtransportSessionWrapper) OpenStreamSync(ctx context.Context) (Stream, error) {
	stream, err := sess.innerSession.OpenStreamSync(ctx)
	return &webtransportStreamWrapper{innerStream: stream}, err
}
func (sess *webtransportSessionWrapper) OpenUniStream() (SendStream, error) {
	stream, err := sess.innerSession.OpenUniStream()
	return &webtransportSendStreamWrapper{innerSendStream: stream}, err
}
func (sess *webtransportSessionWrapper) OpenUniStreamSync(ctx context.Context) (SendStream, error) {
	stream, err := sess.innerSession.OpenUniStreamSync(ctx)
	return &webtransportSendStreamWrapper{innerSendStream: stream}, err
}
func (sess *webtransportSessionWrapper) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return sess.innerSession.ReceiveDatagram(ctx)
}
func (sess *webtransportSessionWrapper) RemoteAddr() net.Addr {
	return sess.innerSession.RemoteAddr()
}
func (sess *webtransportSessionWrapper) SendDatagram(b []byte) error {
	return sess.innerSession.SendDatagram(b)
}
