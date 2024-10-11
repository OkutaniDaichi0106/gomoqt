package moqrawquic

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/quic-go/quic-go"
)

type rawQuicStream struct {
	stream quic.Stream
}

func (wrapper rawQuicStream) StreamID() moqtransport.StreamID {
	return moqtransport.StreamID(wrapper.stream.StreamID())
}

func (wrapper rawQuicStream) Read(b []byte) (int, error) {
	return wrapper.stream.Read(b)
}

func (wrapper rawQuicStream) Write(b []byte) (int, error) {
	return wrapper.stream.Write(b)
}

func (wrapper rawQuicStream) CancelRead(code moqtransport.StreamErrorCode) {
	wrapper.stream.CancelRead(quic.StreamErrorCode(code))
}

func (wrapper rawQuicStream) CancelWrite(code moqtransport.StreamErrorCode) {
	wrapper.stream.CancelWrite(quic.StreamErrorCode(code))
}

func (wrapper rawQuicStream) SetDeadLine(time time.Time) error {
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

type rawQuicReceiveStream struct {
	stream quic.ReceiveStream
}

func (wrapper rawQuicReceiveStream) StreamID() moqtransport.StreamID {
	return moqtransport.StreamID(wrapper.stream.StreamID())
}
func (wrapper rawQuicReceiveStream) Read(b []byte) (int, error) {
	return wrapper.stream.Read(b)
}

func (wrapper rawQuicReceiveStream) CancelRead(code moqtransport.StreamErrorCode) {
	wrapper.stream.CancelRead(quic.StreamErrorCode(code))
}

func (wrapper rawQuicReceiveStream) SetReadDeadline(time time.Time) error {
	return wrapper.stream.SetReadDeadline(time)
}

type rawQuicSendStreamWrapper struct {
	stream quic.SendStream
}

func (wrapper rawQuicSendStreamWrapper) StreamID() moqtransport.StreamID {
	return moqtransport.StreamID(wrapper.stream.StreamID())
}

func (wrapper rawQuicSendStreamWrapper) Write(b []byte) (int, error) {
	return wrapper.stream.Write(b)
}

func (wrapper rawQuicSendStreamWrapper) CancelWrite(code moqtransport.StreamErrorCode) {
	wrapper.stream.CancelWrite(quic.StreamErrorCode(code))
}

func (wrapper rawQuicSendStreamWrapper) SetWriteDeadline(time time.Time) error {
	return wrapper.stream.SetWriteDeadline(time)
}

func (wrapper rawQuicSendStreamWrapper) Close() error {
	return wrapper.stream.Close()
}
