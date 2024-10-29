package moqt

import (
	"io"
	"time"
)

type SendStream interface {
	io.Writer
	io.Closer

	StreamID() StreamID
	CancelWrite(StreamErrorCode)

	SetWriteDeadline(time.Time) error
}

type ReceiveStream interface {
	io.Reader

	StreamID() StreamID
	CancelRead(StreamErrorCode)

	SetReadDeadline(time.Time) error
}

type Stream interface {
	SendStream
	ReceiveStream
	SetDeadLine(time.Time) error
}

type StreamID int64

type StreamErrorCode uint32

type SessionErrorCode uint32

// /*
//  * Handlers
//  */
// type UniStreamHandler interface {
// 	HandleReceiveStream(ReceiveStream)
// 	HandleSendStream(SendStream)
// }

// /***/
// type BiStreamHandler interface {
// 	HandleStream(Stream)
// }

// var _ (BiStreamHandler) = (*defaultBiStreamHandler)(nil)

// type defaultBiStreamHandler struct{}

// func (defaultBiStreamHandler) HandleStream(stream Stream) {

// }

// var _ (UniStreamHandler) = (*defaultUniStreamHandler)(nil)

// type defaultUniStreamHandler struct{}

// func (defaultUniStreamHandler) HandleReceiveStream(stream ReceiveStream) {

// }
// func (defaultUniStreamHandler) HandleSendStream(stream SendStream) {

// }
