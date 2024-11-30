package moq

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

type StreamError interface {
	error
	StreamErrorCode() StreamErrorCode
}
