package quic

import "github.com/quic-go/quic-go"

var ConnectionRefused = quic.ConnectionRefused

type StreamError struct {
	StreamID  StreamID
	ErrorCode StreamErrorCode
	Remote    bool
}

type StreamErrorCode uint32
