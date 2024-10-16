package moqtransport

import (
	"time"

	"github.com/quic-go/quic-go"
)

type rawQuicStream struct {
	stream     quic.Stream
	streamType *StreamType
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
	wrapper.stream.CancelRead(quic.StreamErrorCode(code))
}

func (wrapper rawQuicStream) CancelWrite(code StreamErrorCode) {
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

func (wrapper *rawQuicStream) SetType(streamType StreamType) {
	if wrapper.streamType != nil {
		panic("do not change stream type")
	}

	wrapper.streamType = &streamType
}

func (wrapper rawQuicStream) Type() StreamType {
	return *wrapper.streamType
}

/*
 *
 */
type rawQuicReceiveStream struct {
	stream     quic.ReceiveStream
	streamType *StreamType
}

func (wrapper rawQuicReceiveStream) StreamID() StreamID {
	return StreamID(wrapper.stream.StreamID())
}
func (wrapper rawQuicReceiveStream) Read(b []byte) (int, error) {
	return wrapper.stream.Read(b)
}

func (wrapper rawQuicReceiveStream) CancelRead(code StreamErrorCode) {
	wrapper.stream.CancelRead(quic.StreamErrorCode(code))
}

func (wrapper rawQuicReceiveStream) SetReadDeadline(time time.Time) error {
	return wrapper.stream.SetReadDeadline(time)
}

func (wrapper rawQuicReceiveStream) Type() StreamType {
	return *wrapper.streamType
}

/*
 *
 */
type rawQuicSendStream struct {
	stream     quic.SendStream
	streamType *StreamType
}

func (wrapper rawQuicSendStream) StreamID() StreamID {
	return StreamID(wrapper.stream.StreamID())
}

func (wrapper rawQuicSendStream) Write(b []byte) (int, error) {
	return wrapper.stream.Write(b)
}

func (wrapper rawQuicSendStream) CancelWrite(code StreamErrorCode) {
	wrapper.stream.CancelWrite(quic.StreamErrorCode(code))
}

func (wrapper rawQuicSendStream) SetWriteDeadline(time time.Time) error {
	return wrapper.stream.SetWriteDeadline(time)
}

func (wrapper rawQuicSendStream) Close() error {
	return wrapper.stream.Close()
}

func (wrapper rawQuicSendStream) Type() StreamType {
	return *wrapper.streamType
}