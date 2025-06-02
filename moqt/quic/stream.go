package quic

import (
	"time"
)

type Stream interface {
	SendStream
	ReceiveStream
	SetDeadline(time.Time) error
}

type StreamID uint64
