package moqtransport

import (
	"io"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

type SendByteStream interface {
	io.Writer
	io.Closer

	StreamID() StreamID
	CancelWrite(StreamErrorCode)

	SetWriteDeadline(time.Time) error
}

type ReceiveByteStream interface {
	io.Reader

	StreamID() StreamID
	CancelRead(StreamErrorCode)

	SetReadDeadline(time.Time) error
}

type ByteStream interface {
	SendByteStream
	ReceiveByteStream
	SetDeadLine(time.Time) error
}

type StreamID int64

type StreamErrorCode uint32

type SessionErrorCode uint32 //TODO: move it to session.go

type rawQuicStreamWrapper struct {
	innerStream quic.Stream
}

func (s rawQuicStreamWrapper) StreamID() StreamID {
	return StreamID(s.innerStream.StreamID())
}

func (s rawQuicStreamWrapper) Read(b []byte) (int, error) {
	return s.innerStream.Read(b)
}

func (s rawQuicStreamWrapper) Write(b []byte) (int, error) {
	return s.innerStream.Write(b)
}

func (s rawQuicStreamWrapper) CancelRead(code StreamErrorCode) {
	s.innerStream.CancelRead(quic.StreamErrorCode(code))
}

func (s rawQuicStreamWrapper) CancelWrite(code StreamErrorCode) {
	s.innerStream.CancelWrite(quic.StreamErrorCode(code))
}

func (s rawQuicStreamWrapper) SetDeadLine(time time.Time) error {
	return s.innerStream.SetDeadline(time)
}

func (s rawQuicStreamWrapper) SetReadDeadline(time time.Time) error {
	return s.innerStream.SetReadDeadline(time)
}

func (s rawQuicStreamWrapper) SetWriteDeadline(time time.Time) error {
	return s.innerStream.SetWriteDeadline(time)
}

func (s rawQuicStreamWrapper) Close() error {
	return s.innerStream.Close()
}

type rawQuicReceiveStreamWrapper struct {
	innerReceiveStream quic.ReceiveStream
}

func (s rawQuicReceiveStreamWrapper) StreamID() StreamID {
	return StreamID(s.innerReceiveStream.StreamID())
}
func (s rawQuicReceiveStreamWrapper) Read(b []byte) (int, error) {
	return s.innerReceiveStream.Read(b)
}

func (s rawQuicReceiveStreamWrapper) CancelRead(code StreamErrorCode) {
	s.innerReceiveStream.CancelRead(quic.StreamErrorCode(code))
}

func (s rawQuicReceiveStreamWrapper) SetReadDeadline(time time.Time) error {
	return s.innerReceiveStream.SetReadDeadline(time)
}

type rawQuicSendStreamWrapper struct {
	innerSendStream quic.SendStream
}

func (s rawQuicSendStreamWrapper) StreamID() StreamID {
	return StreamID(s.innerSendStream.StreamID())
}

func (s rawQuicSendStreamWrapper) Write(b []byte) (int, error) {
	return s.innerSendStream.Write(b)
}

func (s rawQuicSendStreamWrapper) CancelWrite(code StreamErrorCode) {
	s.innerSendStream.CancelWrite(quic.StreamErrorCode(code))
}

func (s rawQuicSendStreamWrapper) SetWriteDeadline(time time.Time) error {
	return s.innerSendStream.SetWriteDeadline(time)
}

func (s rawQuicSendStreamWrapper) Close() error {
	return s.innerSendStream.Close()
}

type webtransportStreamWrapper struct {
	innerStream webtransport.Stream
}

func (s webtransportStreamWrapper) StreamID() StreamID {
	return StreamID(s.innerStream.StreamID())
}

func (s webtransportStreamWrapper) Read(b []byte) (int, error) {
	return s.innerStream.Read(b)
}

func (s webtransportStreamWrapper) Write(b []byte) (int, error) {
	return s.innerStream.Write(b)
}

func (s webtransportStreamWrapper) CancelRead(code StreamErrorCode) {
	s.innerStream.CancelRead(webtransport.StreamErrorCode(code))
}

func (s webtransportStreamWrapper) CancelWrite(code StreamErrorCode) {
	s.innerStream.CancelWrite(webtransport.StreamErrorCode(code))
}

func (s webtransportStreamWrapper) SetDeadLine(time time.Time) error {
	return s.innerStream.SetDeadline(time)
}

func (s webtransportStreamWrapper) SetReadDeadline(time time.Time) error {
	return s.innerStream.SetReadDeadline(time)
}

func (s webtransportStreamWrapper) SetWriteDeadline(time time.Time) error {
	return s.innerStream.SetWriteDeadline(time)
}

func (s webtransportStreamWrapper) Close() error {
	return s.innerStream.Close()
}

type webtransportReceiveStreamWrapper struct {
	innerReceiveStream webtransport.ReceiveStream
}

func (s webtransportReceiveStreamWrapper) StreamID() StreamID {
	return StreamID(s.innerReceiveStream.StreamID())
}
func (s webtransportReceiveStreamWrapper) Read(b []byte) (int, error) {
	return s.innerReceiveStream.Read(b)
}

func (s webtransportReceiveStreamWrapper) CancelRead(code StreamErrorCode) {
	s.innerReceiveStream.CancelRead(webtransport.StreamErrorCode(code))
}

func (s webtransportReceiveStreamWrapper) SetReadDeadline(time time.Time) error {
	return s.innerReceiveStream.SetReadDeadline(time)
}

type webtransportSendStreamWrapper struct {
	innerSendStream webtransport.SendStream
}

func (s webtransportSendStreamWrapper) StreamID() StreamID {
	return StreamID(s.innerSendStream.StreamID())
}

func (s webtransportSendStreamWrapper) Write(b []byte) (int, error) {
	return s.innerSendStream.Write(b)
}

func (s webtransportSendStreamWrapper) CancelWrite(code StreamErrorCode) {
	s.innerSendStream.CancelWrite(webtransport.StreamErrorCode(code))
}

func (s webtransportSendStreamWrapper) SetWriteDeadline(time time.Time) error {
	return s.innerSendStream.SetWriteDeadline(time)
}

func (s webtransportSendStreamWrapper) Close() error {
	return s.innerSendStream.Close()
}
