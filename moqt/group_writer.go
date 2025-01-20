package moqt

import (
	"io"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

/*
 * Group Writer
 */
type GroupWriter interface {
	Group
	io.Writer
	WriteFrame([]byte) error
	Close() error
}

/*
 * Group Sender
 */
type SendGroupStream interface {
	GroupWriter

	CancelWrite(StreamErrorCode)
	SetWriteDeadline(time.Time) error
}

var _ SendGroupStream = (*sendGroupStream)(nil)

type sendGroupStream struct {
	Group
	stream transport.SendStream

	startTime time.Time
}

func (stream sendGroupStream) Write(buf []byte) (int, error) {
	return stream.stream.Write(buf)
}

func (stream sendGroupStream) WriteFrame(buf []byte) error {
	fm := message.FrameMessage{
		Payload: buf,
	}
	err := fm.Encode(stream.stream)
	if err != nil {
		return err
	}

	return nil
}

func (stream sendGroupStream) StartAt() time.Time {
	return stream.startTime
}

func (stream sendGroupStream) Close() error {
	return stream.stream.Close()
}

func (stream sendGroupStream) CancelWrite(code StreamErrorCode) {
	stream.stream.CancelWrite(transport.StreamErrorCode(code))
}

func (stream sendGroupStream) SetWriteDeadline(t time.Time) error {
	return stream.stream.SetWriteDeadline(t)
}
