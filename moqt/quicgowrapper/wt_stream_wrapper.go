package quicgowrapper

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/quic-go/webtransport-go"
)

var _ quic.Stream = (*webtransportStream)(nil)

type webtransportStream struct {
	stream webtransport.Stream
}

func (wrapper webtransportStream) StreamID() quic.StreamID {
	return quic.StreamID(wrapper.stream.StreamID())
}

func (wrapper webtransportStream) Read(b []byte) (int, error) {
	return wrapper.stream.Read(b)
}

func (wrapper webtransportStream) Write(b []byte) (int, error) {
	return wrapper.stream.Write(b)
}

func (wrapper webtransportStream) CancelRead(code quic.StreamErrorCode) {
	wrapper.stream.CancelRead(webtransport.StreamErrorCode(code))
}

func (wrapper webtransportStream) CancelWrite(code quic.StreamErrorCode) {
	wrapper.stream.CancelWrite(webtransport.StreamErrorCode(code))
}

func (wrapper webtransportStream) SetDeadline(time time.Time) error {
	return wrapper.stream.SetDeadline(time)
}

func (wrapper webtransportStream) SetReadDeadline(time time.Time) error {
	return wrapper.stream.SetReadDeadline(time)
}

func (wrapper webtransportStream) SetWriteDeadline(time time.Time) error {
	return wrapper.stream.SetWriteDeadline(time)
}

func (wrapper webtransportStream) Close() error {
	return wrapper.stream.Close()
}

/*
 *
 */
type webtransportReceiveStream struct {
	stream webtransport.ReceiveStream
}

func (wrapper webtransportReceiveStream) StreamID() quic.StreamID {
	return quic.StreamID(wrapper.stream.StreamID())
}
func (wrapper webtransportReceiveStream) Read(b []byte) (int, error) {
	return wrapper.stream.Read(b)
}

func (wrapper webtransportReceiveStream) CancelRead(code quic.StreamErrorCode) {
	wrapper.stream.CancelRead(webtransport.StreamErrorCode(code))
}

func (wrapper webtransportReceiveStream) SetReadDeadline(time time.Time) error {
	return wrapper.stream.SetReadDeadline(time)
}

/*
 *
 */
type webtransportSendStream struct {
	stream webtransport.SendStream
}

func (wrapper webtransportSendStream) StreamID() quic.StreamID {
	return quic.StreamID(wrapper.stream.StreamID())
}

func (wrapper webtransportSendStream) Write(b []byte) (int, error) {
	return wrapper.stream.Write(b)
}

func (wrapper webtransportSendStream) CancelWrite(code quic.StreamErrorCode) {
	wrapper.stream.CancelWrite(webtransport.StreamErrorCode(code))
}

func (wrapper webtransportSendStream) SetWriteDeadline(time time.Time) error {
	return wrapper.stream.SetWriteDeadline(time)
}

func (wrapper webtransportSendStream) Close() error {
	return wrapper.stream.Close()
}
