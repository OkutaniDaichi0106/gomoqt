package moqtransport

import (
	"time"

	"github.com/quic-go/webtransport-go"
)

type webtransportStream struct {
	stream     webtransport.Stream
	streamType *StreamType
}

func (wrapper webtransportStream) StreamID() StreamID {
	return StreamID(wrapper.stream.StreamID())
}

func (wrapper webtransportStream) Read(b []byte) (int, error) {
	return wrapper.stream.Read(b)
}

func (wrapper webtransportStream) Write(b []byte) (int, error) {
	return wrapper.stream.Write(b)
}

func (wrapper webtransportStream) CancelRead(code StreamErrorCode) {
	wrapper.stream.CancelRead(webtransport.StreamErrorCode(code))
}

func (wrapper webtransportStream) CancelWrite(code StreamErrorCode) {
	wrapper.stream.CancelWrite(webtransport.StreamErrorCode(code))
}

func (wrapper webtransportStream) SetDeadLine(time time.Time) error {
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

func (wrapper webtransportStream) SetType(streamType StreamType) {
	if wrapper.streamType != nil {
		panic("do not change stream type")
	}

	wrapper.streamType = &streamType
}

func (wrapper webtransportStream) Type() StreamType {
	return *wrapper.streamType
}

/*
 *
 */
type webtransportReceiveStream struct {
	stream     webtransport.ReceiveStream
	streamType *StreamType
}

func (wrapper webtransportReceiveStream) StreamID() StreamID {
	return StreamID(wrapper.stream.StreamID())
}
func (wrapper webtransportReceiveStream) Read(b []byte) (int, error) {
	return wrapper.stream.Read(b)
}

func (wrapper webtransportReceiveStream) CancelRead(code StreamErrorCode) {
	wrapper.stream.CancelRead(webtransport.StreamErrorCode(code))
}

func (wrapper webtransportReceiveStream) SetReadDeadline(time time.Time) error {
	return wrapper.stream.SetReadDeadline(time)
}

func (wrapper webtransportReceiveStream) Type() StreamType {
	return *wrapper.streamType
}

/*
 *
 */
type webtransportSendStream struct {
	stream     webtransport.SendStream
	streamType *StreamType
}

func (wrapper webtransportSendStream) StreamID() StreamID {
	return StreamID(wrapper.stream.StreamID())
}

func (wrapper webtransportSendStream) Write(b []byte) (int, error) {
	return wrapper.stream.Write(b)
}

func (wrapper webtransportSendStream) CancelWrite(code StreamErrorCode) {
	wrapper.stream.CancelWrite(webtransport.StreamErrorCode(code))
}

func (wrapper webtransportSendStream) SetWriteDeadline(time time.Time) error {
	return wrapper.stream.SetWriteDeadline(time)
}

func (wrapper webtransportSendStream) Close() error {
	return wrapper.stream.Close()
}

func (wrapper webtransportSendStream) Type() StreamType {
	return *wrapper.streamType
}
