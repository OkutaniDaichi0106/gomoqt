package moqtransfork

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

type SessionErrorCode uint32 //TODO: move it to session.go

// func ReadStreamType(stream Stream) StreamType {
// 	if stream.StreamType() == nil {
// 		buf := make([]byte, 1)
// 		stream.Read(buf)

// 		return StreamType(buf[0])
// 	}

// 	return *stream.StreamType()
// }
