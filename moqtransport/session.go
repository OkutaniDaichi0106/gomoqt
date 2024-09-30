package moqtransport

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtversion"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/quic-go/webtransport-go"
)

type sessionID uint64

var sessionIDCounter uint64 = 0

func getNextSessionID() sessionID {
	return sessionID(atomic.AddUint64(&sessionIDCounter, 1)) - 1
}

type sessionCore struct {
	sessionID sessionID

	trSess TransportSession

	controlStream ByteStream

	controlReader quicvarint.Reader

	selectedVersion moqtversion.Version
}

func (sess sessionCore) Terminate(err TerminateError) {
	sess.trSess.CloseWithError(SessionErrorCode(err.Code()), err.Error())
}

type AnnounceConfig struct {
	AuthorizationInfo []string

	MaxCacheDuration time.Duration
}

type SubscribeConfig struct {
	moqtmessage.SubscriberPriority
	moqtmessage.GroupOrder

	SubscriptionFilter moqtmessage.SubscriptionFilter

	AuthorizationInfo string
	DeliveryTimeout   time.Duration
}

/*
 * Transport Session: A raw QUIC connection or a WebTransport session
 */
type TransportSession interface {
	AcceptStream(ctx context.Context) (ByteStream, error)
	AcceptUniStream(ctx context.Context) (ReceiveByteStream, error)
	CloseWithError(code SessionErrorCode, msg string) error
	ConnectionState() quic.ConnectionState
	Context() context.Context
	LocalAddr() net.Addr
	OpenStream() (ByteStream, error)
	OpenStreamSync(ctx context.Context) (ByteStream, error)
	OpenUniStream() (SendByteStream, error)
	OpenUniStreamSync(ctx context.Context) (str SendByteStream, err error)
	ReceiveDatagram(ctx context.Context) ([]byte, error)
	RemoteAddr() net.Addr
	SendDatagram(b []byte) error
}

type rawQuicConnectionWrapper struct {
	innerSession quic.Connection
}

func (sess *rawQuicConnectionWrapper) AcceptStream(ctx context.Context) (ByteStream, error) {
	stream, err := sess.innerSession.AcceptStream(ctx)
	return &rawQuicStreamWrapper{innerStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) AcceptUniStream(ctx context.Context) (ReceiveByteStream, error) {
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
func (sess *rawQuicConnectionWrapper) OpenStream() (ByteStream, error) {
	stream, err := sess.innerSession.OpenStream()
	return &rawQuicStreamWrapper{innerStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) OpenStreamSync(ctx context.Context) (ByteStream, error) {
	stream, err := sess.innerSession.OpenStreamSync(ctx)
	return &rawQuicStreamWrapper{innerStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) OpenUniStream() (SendByteStream, error) {
	stream, err := sess.innerSession.OpenUniStream()
	return &rawQuicSendStreamWrapper{innerSendStream: stream}, err
}
func (sess *rawQuicConnectionWrapper) OpenUniStreamSync(ctx context.Context) (SendByteStream, error) {
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

func (sess *webtransportSessionWrapper) AcceptStream(ctx context.Context) (ByteStream, error) {
	stream, err := sess.innerSession.AcceptStream(ctx)
	return &webtransportStreamWrapper{innerStream: stream}, err
}
func (sess *webtransportSessionWrapper) AcceptUniStream(ctx context.Context) (ReceiveByteStream, error) {
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
func (sess *webtransportSessionWrapper) OpenStream() (ByteStream, error) {
	stream, err := sess.innerSession.OpenStream()
	return &webtransportStreamWrapper{innerStream: stream}, err
}
func (sess *webtransportSessionWrapper) OpenStreamSync(ctx context.Context) (ByteStream, error) {
	stream, err := sess.innerSession.OpenStreamSync(ctx)
	return &webtransportStreamWrapper{innerStream: stream}, err
}
func (sess *webtransportSessionWrapper) OpenUniStream() (SendByteStream, error) {
	stream, err := sess.innerSession.OpenUniStream()
	return &webtransportSendStreamWrapper{innerSendStream: stream}, err
}
func (sess *webtransportSessionWrapper) OpenUniStreamSync(ctx context.Context) (SendByteStream, error) {
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
