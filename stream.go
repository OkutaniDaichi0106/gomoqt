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

const (
	stream_internal_error StreamErrorCode = 0x00
	invalid_stream_type   StreamErrorCode = 0x10 // TODO: See spec
)

type StreamError interface {
	error
	StreamErrorCode() StreamErrorCode
}

type defaultStreamError struct {
	code   StreamErrorCode
	reason string
}

func (err defaultStreamError) Error() string {
	return err.reason
}

func (err defaultStreamError) StreamErrorCode() StreamErrorCode {
	return err.code
}

var (
	ErrInvalidStreamType = defaultStreamError{
		code:   invalid_stream_type,
		reason: "invalid stream type",
	}
)
