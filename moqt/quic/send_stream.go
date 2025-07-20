package quic

import (
	"context"
	"io"
	"time"
)

type SendStream interface {
	io.Writer
	io.Closer

	StreamID() StreamID
	CancelWrite(StreamErrorCode)

	SetWriteDeadline(time.Time) error

	Context() context.Context
}
