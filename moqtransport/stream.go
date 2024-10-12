package moqtransport

import (
	"io"
	"time"
)

type StreamType uint8

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

	// moqt
	SetType(StreamType)
	Type() StreamType
}

type StreamID int64

type StreamErrorCode uint32

type SessionErrorCode uint32 //TODO: move it to session.go
