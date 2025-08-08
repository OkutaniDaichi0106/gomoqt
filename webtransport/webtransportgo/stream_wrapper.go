package webtransportgo

import (
	"context"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
	quicgo_webtransportgo "github.com/OkutaniDaichi0106/webtransport-go"
)

var _ quic.Stream = (*streamWrapper)(nil)

type streamWrapper struct {
	stream *quicgo_webtransportgo.Stream
}

func (wrapper streamWrapper) StreamID() quic.StreamID {
	return quic.StreamID(wrapper.stream.StreamID())
}

func (wrapper streamWrapper) Read(b []byte) (int, error) {
	n, err := wrapper.stream.Read(b)
	return n, wrapError(err)
}

func (wrapper streamWrapper) Write(b []byte) (int, error) {
	n, err := wrapper.stream.Write(b)
	return n, wrapError(err)
}

func (wrapper streamWrapper) CancelRead(code quic.StreamErrorCode) {
	wrapper.stream.CancelRead(quicgo_webtransportgo.StreamErrorCode(code))
}

func (wrapper streamWrapper) CancelWrite(code quic.StreamErrorCode) {
	wrapper.stream.CancelWrite(quicgo_webtransportgo.StreamErrorCode(code))
}

func (wrapper streamWrapper) SetDeadline(time time.Time) error {
	return wrapError(wrapper.stream.SetDeadline(time))
}

func (wrapper streamWrapper) SetReadDeadline(time time.Time) error {
	return wrapError(wrapper.stream.SetReadDeadline(time))
}

func (wrapper streamWrapper) SetWriteDeadline(time time.Time) error {
	return wrapError(wrapper.stream.SetWriteDeadline(time))
}

func (wrapper streamWrapper) Close() error {
	return wrapError(wrapper.stream.Close())
}

func (wrapper streamWrapper) Context() context.Context {
	return wrapper.stream.Context()
}

/*
 *
 */
type receiveStreamWrapper struct {
	stream *quicgo_webtransportgo.ReceiveStream
}

func (wrapper receiveStreamWrapper) StreamID() quic.StreamID {
	return quic.StreamID(wrapper.stream.StreamID())
}
func (wrapper receiveStreamWrapper) Read(b []byte) (int, error) {
	n, err := wrapper.stream.Read(b)
	return n, wrapError(err)
}

func (wrapper receiveStreamWrapper) CancelRead(code quic.StreamErrorCode) {
	wrapper.stream.CancelRead(quicgo_webtransportgo.StreamErrorCode(code))
}

func (wrapper receiveStreamWrapper) SetReadDeadline(time time.Time) error {
	return wrapError(wrapper.stream.SetReadDeadline(time))
}

/*
 *
 */
type sendStreamWrapper struct {
	stream *quicgo_webtransportgo.SendStream
}

func (wrapper sendStreamWrapper) StreamID() quic.StreamID {
	return quic.StreamID(wrapper.stream.StreamID())
}

func (wrapper sendStreamWrapper) Write(b []byte) (int, error) {
	n, err := wrapper.stream.Write(b)
	return n, wrapError(err)
}

func (wrapper sendStreamWrapper) CancelWrite(code quic.StreamErrorCode) {
	wrapper.stream.CancelWrite(quicgo_webtransportgo.StreamErrorCode(code))
}

func (wrapper sendStreamWrapper) SetWriteDeadline(time time.Time) error {
	return wrapError(wrapper.stream.SetWriteDeadline(time))
}

func (wrapper sendStreamWrapper) Close() error {
	return wrapError(wrapper.stream.Close())
}

func (wrapper sendStreamWrapper) Context() context.Context {
	return wrapper.stream.Context()
}
