package quicgo

import (
	"context"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	quicgo_quicgo "github.com/quic-go/quic-go"
)

var _ quic.Stream = (*rawQuicStream)(nil)

type rawQuicStream struct {
	stream *quicgo_quicgo.Stream
}

func (wrapper rawQuicStream) StreamID() quic.StreamID {
	return quic.StreamID(wrapper.stream.StreamID())
}

func (wrapper rawQuicStream) Read(b []byte) (int, error) {
	n, err := wrapper.stream.Read(b)
	return n, WrapError(err)
}

func (wrapper rawQuicStream) Write(b []byte) (int, error) {
	n, err := wrapper.stream.Write(b)
	return n, WrapError(err)
}

func (wrapper rawQuicStream) CancelRead(code quic.StreamErrorCode) {
	wrapper.stream.CancelRead(quicgo_quicgo.StreamErrorCode(code))
}

func (wrapper rawQuicStream) CancelWrite(code quic.StreamErrorCode) {
	wrapper.stream.CancelWrite(quicgo_quicgo.StreamErrorCode(code))
}

func (wrapper rawQuicStream) SetDeadline(time time.Time) error {
	return WrapError(wrapper.stream.SetDeadline(time))
}

func (wrapper rawQuicStream) SetReadDeadline(time time.Time) error {
	return WrapError(wrapper.stream.SetReadDeadline(time))
}

func (wrapper rawQuicStream) SetWriteDeadline(time time.Time) error {
	return WrapError(wrapper.stream.SetWriteDeadline(time))
}

func (wrapper rawQuicStream) Close() error {
	return WrapError(wrapper.stream.Close())
}

func (wrapper rawQuicStream) Context() context.Context {
	return wrapper.stream.Context()
}

/*
 *
 */
var _ quic.ReceiveStream = (*rawQuicReceiveStream)(nil)

type rawQuicReceiveStream struct {
	stream *quicgo_quicgo.ReceiveStream
}

func (wrapper rawQuicReceiveStream) StreamID() quic.StreamID {
	return quic.StreamID(wrapper.stream.StreamID())
}
func (wrapper rawQuicReceiveStream) Read(b []byte) (int, error) {
	n, err := wrapper.stream.Read(b)
	return n, WrapError(err)
}

func (wrapper rawQuicReceiveStream) CancelRead(code quic.StreamErrorCode) {
	wrapper.stream.CancelRead(quicgo_quicgo.StreamErrorCode(code))
}

func (wrapper rawQuicReceiveStream) SetReadDeadline(time time.Time) error {
	return WrapError(wrapper.stream.SetReadDeadline(time))
}

/*
 *
 */

var _ quic.SendStream = (*rawQuicSendStream)(nil)

type rawQuicSendStream struct {
	stream *quicgo_quicgo.SendStream
}

func (wrapper rawQuicSendStream) StreamID() quic.StreamID {
	return quic.StreamID(wrapper.stream.StreamID())
}

func (wrapper rawQuicSendStream) Write(b []byte) (int, error) {
	n, err := wrapper.stream.Write(b)
	return n, WrapError(err)
}

func (wrapper rawQuicSendStream) CancelWrite(code quic.StreamErrorCode) {
	wrapper.stream.CancelWrite(quicgo_quicgo.StreamErrorCode(code))
}

func (wrapper rawQuicSendStream) SetWriteDeadline(time time.Time) error {
	return WrapError(wrapper.stream.SetWriteDeadline(time))
}

func (wrapper rawQuicSendStream) Close() error {
	return WrapError(wrapper.stream.Close())
}

func (wrapper rawQuicSendStream) Context() context.Context {
	return wrapper.stream.Context()
}
