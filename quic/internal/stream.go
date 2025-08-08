package internal

import (
	"context"
	"io"
	"time"
)

type Stream interface {
	SendStream
	ReceiveStream
	SetDeadline(time.Time) error
}
type SendStream interface {
	io.Writer
	io.Closer

	StreamID() StreamID
	CancelWrite(StreamErrorCode)

	SetWriteDeadline(time.Time) error

	Context() context.Context
}
type ReceiveStream interface {
	io.Reader

	StreamID() StreamID

	CancelRead(StreamErrorCode)

	SetReadDeadline(time.Time) error
}
type StreamID uint64
