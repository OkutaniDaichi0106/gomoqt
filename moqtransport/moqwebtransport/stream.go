package moqwebtransport

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport"
	"github.com/quic-go/webtransport-go"
)

type Stream interface {
	moqtransport.Stream
}

type webtransportStream struct {
	str webtransport.Stream
}

func (wts webtransportStream) StreamID() moqtransport.StreamID {
	return moqtransport.StreamID(wts.str.StreamID())
}

func (wts webtransportStream) Read(b []byte) (int, error) {
	return wts.str.Read(b)
}

func (wts webtransportStream) Write(b []byte) (int, error) {
	return wts.str.Write(b)
}

func (wts webtransportStream) CancelRead(code moqtransport.StreamErrorCode) {
	wts.str.CancelRead(webtransport.StreamErrorCode(code))
}

func (wts webtransportStream) CancelWrite(code moqtransport.StreamErrorCode) {
	wts.str.CancelWrite(webtransport.StreamErrorCode(code))
}

func (wts webtransportStream) SetDeadLine(time time.Time) error {
	return wts.str.SetDeadline(time)
}

func (wts webtransportStream) SetReadDeadline(time time.Time) error {
	return wts.str.SetReadDeadline(time)
}

func (wts webtransportStream) SetWriteDeadline(time time.Time) error {
	return wts.str.SetWriteDeadline(time)
}

func (wts webtransportStream) Close() error {
	return wts.str.Close()
}

type webtransportReceiveStream struct {
	innerReceiveStream webtransport.ReceiveStream
}

func (wts webtransportReceiveStream) StreamID() moqtransport.StreamID {
	return moqtransport.StreamID(wts.innerReceiveStream.StreamID())
}
func (wts webtransportReceiveStream) Read(b []byte) (int, error) {
	return wts.innerReceiveStream.Read(b)
}

func (wts webtransportReceiveStream) CancelRead(code moqtransport.StreamErrorCode) {
	wts.innerReceiveStream.CancelRead(webtransport.StreamErrorCode(code))
}

func (wts webtransportReceiveStream) SetReadDeadline(time time.Time) error {
	return wts.innerReceiveStream.SetReadDeadline(time)
}

type webtransportSendStream struct {
	innerSendStream webtransport.SendStream
}

func (wts webtransportSendStream) StreamID() moqtransport.StreamID {
	return moqtransport.StreamID(wts.innerSendStream.StreamID())
}

func (wts webtransportSendStream) Write(b []byte) (int, error) {
	return wts.innerSendStream.Write(b)
}

func (wts webtransportSendStream) CancelWrite(code moqtransport.StreamErrorCode) {
	wts.innerSendStream.CancelWrite(webtransport.StreamErrorCode(code))
}

func (wts webtransportSendStream) SetWriteDeadline(time time.Time) error {
	return wts.innerSendStream.SetWriteDeadline(time)
}

func (wts webtransportSendStream) Close() error {
	return wts.innerSendStream.Close()
}
