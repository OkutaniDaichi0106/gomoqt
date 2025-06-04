package webtransportgo

import (
	"context"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo_quicgo "github.com/quic-go/quic-go"
)

func UnwrapStream(stream quic.Stream) quicgo_quicgo.Stream {
	if stream == nil {
		return nil
	}
	return &quicgoStreamWrapper{stream: stream}
}

var _ quicgo_quicgo.Stream = (*quicgoStreamWrapper)(nil)

type quicgoStreamWrapper struct {
	stream quic.Stream
}

func (unwrapper *quicgoStreamWrapper) StreamID() quicgo_quicgo.StreamID {
	return quicgo_quicgo.StreamID(unwrapper.stream.StreamID())
}

func (wrapper *quicgoStreamWrapper) Read(b []byte) (int, error) {
	return wrapper.stream.Read(b)
}

func (wrapper *quicgoStreamWrapper) Write(b []byte) (int, error) {
	return wrapper.stream.Write(b)
}
func (wrapper *quicgoStreamWrapper) CancelRead(code quicgo_quicgo.StreamErrorCode) {
	wrapper.stream.CancelRead(quic.StreamErrorCode(code))
}

func (wrapper *quicgoStreamWrapper) CancelWrite(code quicgo_quicgo.StreamErrorCode) {
	wrapper.stream.CancelWrite(quic.StreamErrorCode(code))
}

func (wrapper *quicgoStreamWrapper) SetDeadline(time time.Time) error {
	return wrapper.stream.SetDeadline(time)
}

func (wrapper *quicgoStreamWrapper) SetReadDeadline(time time.Time) error {
	return wrapper.stream.SetReadDeadline(time)
}

func (wrapper *quicgoStreamWrapper) SetWriteDeadline(time time.Time) error {
	return wrapper.stream.SetWriteDeadline(time)
}

func (wrapper *quicgoStreamWrapper) Close() error {
	return wrapper.stream.Close()
}

func (wrapper *quicgoStreamWrapper) Context() context.Context {
	return context.Background()
}

// /
func UnwrapReceiveStream(stream quic.ReceiveStream) quicgo_quicgo.ReceiveStream {
	if stream == nil {
		return nil
	}
	return &quicgoReceiveStreamWrapper{stream: stream}
}

var _ quicgo_quicgo.ReceiveStream = (*quicgoReceiveStreamWrapper)(nil)

type quicgoReceiveStreamWrapper struct {
	stream quic.ReceiveStream
}

func (wrapper *quicgoReceiveStreamWrapper) StreamID() quicgo_quicgo.StreamID {
	return quicgo_quicgo.StreamID(wrapper.stream.StreamID())
}

func (wrapper *quicgoReceiveStreamWrapper) Read(b []byte) (int, error) {
	return wrapper.stream.Read(b)
}

func (wrapper *quicgoReceiveStreamWrapper) CancelRead(code quicgo_quicgo.StreamErrorCode) {
	wrapper.stream.CancelRead(quic.StreamErrorCode(code))
}

func (wrapper *quicgoReceiveStreamWrapper) SetReadDeadline(time time.Time) error {
	return wrapper.stream.SetReadDeadline(time)
}

var _ quicgo_quicgo.SendStream = (*quicgoSendStreamWrapper)(nil)

type quicgoSendStreamWrapper struct {
	stream quic.SendStream
}

func (wrapper *quicgoSendStreamWrapper) StreamID() quicgo_quicgo.StreamID {
	return quicgo_quicgo.StreamID(wrapper.stream.StreamID())
}
func (wrapper *quicgoSendStreamWrapper) Write(b []byte) (int, error) {
	return wrapper.stream.Write(b)
}
func (wrapper *quicgoSendStreamWrapper) CancelWrite(code quicgo_quicgo.StreamErrorCode) {
	wrapper.stream.CancelWrite(quic.StreamErrorCode(code))
}
func (wrapper *quicgoSendStreamWrapper) SetWriteDeadline(time time.Time) error {
	return wrapper.stream.SetWriteDeadline(time)
}
func (wrapper *quicgoSendStreamWrapper) Close() error {
	return wrapper.stream.Close()
}
func (wrapper *quicgoSendStreamWrapper) Context() context.Context {
	return context.Background()
}
