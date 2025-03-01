package internal

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

func newReceiveGroupStream(gm *message.GroupMessage, stream transport.ReceiveStream) *ReceiveGroupStream {
	return &ReceiveGroupStream{
		GroupMessage:  *gm,
		ReceiveStream: stream,
		startTime:     time.Now(),
	}
}

type ReceiveGroupStream struct {
	GroupMessage  message.GroupMessage
	ReceiveStream transport.ReceiveStream

	startTime time.Time

	//errCodeCh chan StreamErrorCode
}

func (r ReceiveGroupStream) ReadFrameBytes() ([]byte, error) {
	var fm message.FrameMessage
	_, err := fm.Decode(r.ReceiveStream)
	if err != nil {
		return nil, err
	}

	return fm.Payload, nil
}

func (r ReceiveGroupStream) StartAt() time.Time {
	return r.startTime
}

func (r ReceiveGroupStream) CancelRead(code protocol.GroupErrorCode) {
	r.ReceiveStream.CancelRead(transport.StreamErrorCode(code))
}

func (r ReceiveGroupStream) SetReadDeadline(t time.Time) error {
	return r.ReceiveStream.SetReadDeadline(t)
}
