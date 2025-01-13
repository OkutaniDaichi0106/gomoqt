package moqt

import (
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type SendDataStream interface {
	transport.SendStream
	SentGroup
}

var _ SendDataStream = (*sendDataStream)(nil)

type sendDataStream struct {
	transport.SendStream
	sentGroup
}

func (stream sendDataStream) Write(buf []byte) (int, error) {
	fm := message.FrameMessage{
		Payload: buf,
	}
	err := fm.Encode(stream.SendStream)
	if err != nil {
		return 0, err
	}

	return len(buf), nil
}

// var _ GroupReader = (ReceiveDataStream)(nil)

type ReceiveDataStream interface {
	SubscribeID() SubscribeID
	transport.ReceiveStream
	NextFrame() ([]byte, error)
	ReceivedGroup
}

var _ ReceiveDataStream = (*receiveDataStream)(nil)

type receiveDataStream struct {
	subscribeID SubscribeID
	transport.ReceiveStream
	ReceivedGroup
}

func (stream receiveDataStream) SubscribeID() SubscribeID {
	return stream.subscribeID
}

func (stream receiveDataStream) Read(buf []byte) (int, error) {
	var fm message.FrameMessage
	err := fm.Decode(stream.ReceiveStream)
	if err != nil {
		return 0, err
	}

	n := copy(buf, fm.Payload)

	return n, nil
}

func (stream receiveDataStream) NextFrame() ([]byte, error) {
	var fm message.FrameMessage
	err := fm.Decode(stream.ReceiveStream)
	if err != nil {
		return nil, err
	}

	return fm.Payload, nil
}

func newReceiveDataStreamQueue() *receiveDataStreamQueue {
	return &receiveDataStreamQueue{
		queue: make([]ReceiveDataStream, 0), // TODO: Tune the initial capacity
		ch:    make(chan struct{}, 1),
	}
}

type receiveDataStreamQueue struct {
	queue []ReceiveDataStream
	ch    chan struct{}
	mu    sync.Mutex
}

func (q *receiveDataStreamQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *receiveDataStreamQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *receiveDataStreamQueue) Enqueue(stream ReceiveDataStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, stream)
}

func (q *receiveDataStreamQueue) Dequeue() ReceiveDataStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	next := q.queue[0]

	q.queue = q.queue[1:]

	return next
}
