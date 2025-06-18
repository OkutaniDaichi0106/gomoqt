package webtransportgo

import (
	"context"
	"errors"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo_quicgo "github.com/quic-go/quic-go"
)

func UnwrapConnection(conn quic.Connection) quicgo_quicgo.Connection {
	if conn == nil {
		return nil
	}
	return &quicgoConnection{conn: conn}
}

var _ quicgo_quicgo.Connection = (*quicgoConnection)(nil)

type quicgoConnection struct {
	conn quic.Connection
}

func (c *quicgoConnection) OpenStream() (quicgo_quicgo.Stream, error) {
	stream, err := c.conn.OpenStream()
	if err != nil {
		return nil, err
	}
	return &quicgoStreamWrapper{stream: stream}, nil
}

func (c *quicgoConnection) OpenStreamSync(ctx context.Context) (quicgo_quicgo.Stream, error) {
	stream, err := c.conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	return &quicgoStreamWrapper{stream: stream}, nil
}

func (c *quicgoConnection) OpenUniStream() (quicgo_quicgo.SendStream, error) {
	stream, err := c.conn.OpenUniStream()
	if err != nil {
		return nil, err
	}
	return &quicgoSendStreamWrapper{stream: stream}, nil
}

func (c *quicgoConnection) OpenUniStreamSync(ctx context.Context) (quicgo_quicgo.SendStream, error) {
	stream, err := c.conn.OpenUniStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	return &quicgoSendStreamWrapper{stream: stream}, nil
}

func (c *quicgoConnection) AcceptStream(ctx context.Context) (quicgo_quicgo.Stream, error) {
	stream, err := c.conn.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}
	return &quicgoStreamWrapper{stream: stream}, nil
}

func (c *quicgoConnection) AcceptUniStream(ctx context.Context) (quicgo_quicgo.ReceiveStream, error) {
	stream, err := c.conn.AcceptUniStream(ctx)
	if err != nil {
		return nil, err
	}
	return &quicgoReceiveStreamWrapper{stream: stream}, nil

}

func (c *quicgoConnection) CloseWithError(code quicgo_quicgo.ApplicationErrorCode, msg string) error {
	return c.conn.CloseWithError(quic.ConnectionErrorCode(code), msg)
}

func (c *quicgoConnection) ConnectionState() quicgo_quicgo.ConnectionState {
	state := c.conn.ConnectionState()

	return quicgo_quicgo.ConnectionState{
		TLS:               state.TLS,
		SupportsDatagrams: state.SupportsDatagrams,
		Used0RTT:          state.Used0RTT,
		Version:           quicgo_quicgo.Version(state.Version),
		GSO:               state.GSO,
	}
}

func (c *quicgoConnection) Context() context.Context {
	return c.conn.Context()
}

func (c *quicgoConnection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *quicgoConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *quicgoConnection) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return nil, errors.New("not supported")
}

func (c *quicgoConnection) SendDatagram(data []byte) error {
	return errors.New("not supported")
}

func (c *quicgoConnection) AddPath(*quicgo_quicgo.Transport) (*quicgo_quicgo.Path, error) {
	return nil, errors.New("not supported")
}
