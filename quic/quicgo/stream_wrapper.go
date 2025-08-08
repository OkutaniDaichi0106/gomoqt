package quicgo

import (
	"context"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/quic/internal"
	quicgo_quicgo "github.com/quic-go/quic-go"
)

var _ internal.Stream = (*rawQuicStream)(nil)

type rawQuicStream struct {
	stream *quicgo_quicgo.Stream
}

func (wrapper rawQuicStream) StreamID() internal.StreamID {
	return internal.StreamID(wrapper.stream.StreamID())
}

func (wrapper rawQuicStream) Read(b []byte) (int, error) {
	n, err := wrapper.stream.Read(b)
	return n, wrapError(err)
}

func (wrapper rawQuicStream) Write(b []byte) (int, error) {
	n, err := wrapper.stream.Write(b)
	return n, wrapError(err)
}

func (wrapper rawQuicStream) CancelRead(code internal.StreamErrorCode) {
	wrapper.stream.CancelRead(quicgo_quicgo.StreamErrorCode(code))
}

func (wrapper rawQuicStream) CancelWrite(code internal.StreamErrorCode) {
	wrapper.stream.CancelWrite(quicgo_quicgo.StreamErrorCode(code))
}

func (wrapper rawQuicStream) SetDeadline(time time.Time) error {
	return wrapError(wrapper.stream.SetDeadline(time))
}

func (wrapper rawQuicStream) SetReadDeadline(time time.Time) error {
	return wrapError(wrapper.stream.SetReadDeadline(time))
}

func (wrapper rawQuicStream) SetWriteDeadline(time time.Time) error {
	return wrapError(wrapper.stream.SetWriteDeadline(time))
}

func (wrapper rawQuicStream) Close() error {
	return wrapError(wrapper.stream.Close())
}

func (wrapper rawQuicStream) Context() context.Context {
	return wrapper.stream.Context()
}

/*
 *
 */
var _ internal.ReceiveStream = (*rawQuicReceiveStream)(nil)

type rawQuicReceiveStream struct {
	stream *quicgo_quicgo.ReceiveStream
}

func (wrapper rawQuicReceiveStream) StreamID() internal.StreamID {
	return internal.StreamID(wrapper.stream.StreamID())
}
func (wrapper rawQuicReceiveStream) Read(b []byte) (int, error) {
	n, err := wrapper.stream.Read(b)
	return n, wrapError(err)
}

func (wrapper rawQuicReceiveStream) CancelRead(code internal.StreamErrorCode) {
	wrapper.stream.CancelRead(quicgo_quicgo.StreamErrorCode(code))
}

func (wrapper rawQuicReceiveStream) SetReadDeadline(time time.Time) error {
	return wrapError(wrapper.stream.SetReadDeadline(time))
}

/*
 *
 */

var _ internal.SendStream = (*rawQuicSendStream)(nil)

type rawQuicSendStream struct {
	stream *quicgo_quicgo.SendStream
}

func (wrapper rawQuicSendStream) StreamID() internal.StreamID {
	return internal.StreamID(wrapper.stream.StreamID())
}

func (wrapper rawQuicSendStream) Write(b []byte) (int, error) {
	n, err := wrapper.stream.Write(b)
	return n, wrapError(err)
}

func (wrapper rawQuicSendStream) CancelWrite(code internal.StreamErrorCode) {
	wrapper.stream.CancelWrite(quicgo_quicgo.StreamErrorCode(code))
}

func (wrapper rawQuicSendStream) SetWriteDeadline(time time.Time) error {
	return wrapError(wrapper.stream.SetWriteDeadline(time))
}

func (wrapper rawQuicSendStream) Close() error {
	return wrapError(wrapper.stream.Close())
}

func (wrapper rawQuicSendStream) Context() context.Context {
	return wrapper.stream.Context()
}
