package quicgowrapper

import (
	"context"
	"errors"
	"net"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo "github.com/quic-go/quic-go"
)

var _ quicgo.Connection = (*quicgoConnection)(nil)

type quicgoConnection struct {
	conn quic.Connection
}

func (c *quicgoConnection) OpenStream() (quicgo.Stream, error) {
	stream, err := c.conn.OpenStream()
	if err != nil {
		return nil, err
	}
	return &quicgoStreamWrapper{stream: stream}, nil
}

func (c *quicgoConnection) OpenStreamSync(ctx context.Context) (quicgo.Stream, error) {
	stream, err := c.conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	return &quicgoStreamWrapper{stream: stream}, nil
}

func (c *quicgoConnection) OpenUniStream() (quicgo.SendStream, error) {
	stream, err := c.conn.OpenUniStream()
	if err != nil {
		return nil, err
	}
	return &quicgoSendStreamWrapper{stream: stream}, nil
}

func (c *quicgoConnection) OpenUniStreamSync(ctx context.Context) (quicgo.SendStream, error) {
	stream, err := c.conn.OpenUniStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	return &quicgoSendStreamWrapper{stream: stream}, nil
}

func (c *quicgoConnection) AcceptStream(ctx context.Context) (quicgo.Stream, error) {
	stream, err := c.conn.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}
	return &quicgoStreamWrapper{stream: stream}, nil
}

func (c *quicgoConnection) AcceptUniStream(ctx context.Context) (quicgo.ReceiveStream, error) {
	stream, err := c.conn.AcceptUniStream(ctx)
	if err != nil {
		return nil, err
	}
	return &quicgoReceiveStreamWrapper{stream: stream}, nil

}

func (c *quicgoConnection) CloseWithError(code quicgo.ApplicationErrorCode, msg string) error {
	return c.conn.CloseWithError(quic.ConnectionErrorCode(code), msg)
}

func (c *quicgoConnection) ConnectionState() quicgo.ConnectionState {
	return quicgo.ConnectionState(c.conn.ConnectionState())
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

// func (c *quicgoConnection) AddPath(*quicgo.Transport) (*quicgo.Path, error) {
// 	return nil, errors.New("not supported")
// }
