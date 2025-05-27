package quic

import (
	"context"
	"time"

	quicgo "github.com/quic-go/quic-go"
)

var _ Stream = (*rawQuicStream)(nil)

type rawQuicStream struct {
	stream quicgo.Stream
}

func (wrapper rawQuicStream) StreamID() StreamID {
	return StreamID(wrapper.stream.StreamID())
}

func (wrapper rawQuicStream) Read(b []byte) (int, error) {
	return wrapper.stream.Read(b)
}

func (wrapper rawQuicStream) Write(b []byte) (int, error) {
	return wrapper.stream.Write(b)
}

func (wrapper rawQuicStream) CancelRead(code StreamErrorCode) {
	wrapper.stream.CancelRead(quicgo.StreamErrorCode(code))
}

func (wrapper rawQuicStream) CancelWrite(code StreamErrorCode) {
	wrapper.stream.CancelWrite(quicgo.StreamErrorCode(code))
}

func (wrapper rawQuicStream) SetDeadline(time time.Time) error {
	return wrapper.stream.SetDeadline(time)
}

func (wrapper rawQuicStream) SetReadDeadline(time time.Time) error {
	return wrapper.stream.SetReadDeadline(time)
}

func (wrapper rawQuicStream) SetWriteDeadline(time time.Time) error {
	return wrapper.stream.SetWriteDeadline(time)
}

func (wrapper rawQuicStream) Close() error {
	return wrapper.stream.Close()
}

func (wrapper rawQuicStream) Context() context.Context {
	return wrapper.stream.Context()
}

/*
 *
 */
var _ ReceiveStream = (*rawQuicReceiveStream)(nil)

type rawQuicReceiveStream struct {
	stream quicgo.ReceiveStream
}

func (wrapper rawQuicReceiveStream) StreamID() StreamID {
	return StreamID(wrapper.stream.StreamID())
}
func (wrapper rawQuicReceiveStream) Read(b []byte) (int, error) {
	return wrapper.stream.Read(b)
}

func (wrapper rawQuicReceiveStream) CancelRead(code StreamErrorCode) {
	wrapper.stream.CancelRead(quicgo.StreamErrorCode(code))
}

func (wrapper rawQuicReceiveStream) SetReadDeadline(time time.Time) error {
	return wrapper.stream.SetReadDeadline(time)
}

/*
 *
 */

var _ SendStream = (*rawQuicSendStream)(nil)

type rawQuicSendStream struct {
	stream quicgo.SendStream
}

func (wrapper rawQuicSendStream) StreamID() StreamID {
	return StreamID(wrapper.stream.StreamID())
}

func (wrapper rawQuicSendStream) Write(b []byte) (int, error) {
	return wrapper.stream.Write(b)
}

func (wrapper rawQuicSendStream) CancelWrite(code StreamErrorCode) {
	wrapper.stream.CancelWrite(quicgo.StreamErrorCode(code))
}

func (wrapper rawQuicSendStream) SetWriteDeadline(time time.Time) error {
	return wrapper.stream.SetWriteDeadline(time)
}

func (wrapper rawQuicSendStream) Close() error {
	return wrapper.stream.Close()
}

func (wrapper rawQuicSendStream) Context() context.Context {
	return wrapper.stream.Context()
}
