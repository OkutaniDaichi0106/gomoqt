package quic

import (
	"time"
)

type Stream interface {
	SendStream
	ReceiveStream
	SetDeadline(time.Time) error
}

type StreamID int64

type StreamErrorCode uint32

type StreamError interface {
	error
	StreamErrorCode() StreamErrorCode
}
