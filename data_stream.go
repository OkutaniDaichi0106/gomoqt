package moqt

import (
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type DataSendStream interface {
	transport.SendStream
	SentGroup
}

var _ DataSendStream = (*dataSendStream)(nil)

type dataSendStream struct {
	transport.SendStream
	sentGroup
}

func (stream dataSendStream) Write(buf []byte) (int, error) {
	fm := message.FrameMessage{
		Payload: buf,
	}
	err := fm.Encode(stream.SendStream)
	if err != nil {
		return 0, err
	}

	return len(buf), nil
}

type DataReceiveStream interface {
	transport.ReceiveStream
	ReceivedGroup
}

func newDataReceiveStream(stream transport.ReceiveStream) (DataReceiveStream, error) {
	group, err := readGroup(stream)
	if err != nil {
		slog.Error("failed to get a group", slog.String("error", err.Error()))
		return nil, err
	}

	return &dataReceiveStream{
		ReceiveStream: stream,
		receivedGroup: group,
	}, nil
}

var _ DataReceiveStream = (*dataReceiveStream)(nil)

type dataReceiveStream struct {
	transport.ReceiveStream
	receivedGroup
}

func (stream dataReceiveStream) Read(buf []byte) (int, error) {
	var fm message.FrameMessage
	err := fm.Decode(stream.ReceiveStream)
	if err != nil {
		return 0, err
	}

	n := copy(buf, fm.Payload)

	return n, nil
}

type dataReceiveStreamQueue struct {
	queue []DataReceiveStream
	ch    chan struct{}
	mu    sync.Mutex
}

func (q *dataReceiveStreamQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *dataReceiveStreamQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *dataReceiveStreamQueue) Enqueue(stream DataReceiveStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, stream)
}

func (q *dataReceiveStreamQueue) Dequeue() DataReceiveStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	next := q.queue[0]

	q.queue = q.queue[1:]

	return next
}
