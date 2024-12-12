package moqt

import (
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

var _ DataSender = (*dataSender)(nil)

type DataSender interface {
}

type dataSender struct{}

/*
 *
 */
type dataReceiverQueue struct {
	queue []DataReceiver
}

var _ DataReceiver = (*dataReceiver)(nil)

type DataReceiver interface {
	io.Reader
	GroupSequence() GroupSequence
	SubscribeID() SubscribeID
}

type dataReceiver struct {
	Group
	stream moq.ReceiveStream
}

func (dr dataReceiver) Read(buf []byte) (int, error) {
	var fm message.FrameMessage
	err := fm.Decode(dr.stream)
	if err != nil {
		return 0, err
	}

	n := copy(buf, fm.Payload)

	return n, nil
}
