package quic

import (
	"io"
	"time"
)

type ReceiveStream interface {
	io.Reader

	StreamID() StreamID

	CancelRead(StreamErrorCode)

	SetReadDeadline(time.Time) error
}
