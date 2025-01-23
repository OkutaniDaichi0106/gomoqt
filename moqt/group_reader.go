package moqt

import (
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

/*
 * Group Reader
 */
type GroupReader interface {
	GroupSequence() GroupSequence
	ReadFrame() ([]byte, error)
}

type ReceiveGroupStream interface {
	GroupReader

	SubscribeID() SubscribeID

	CancelRead(StreamErrorCode)

	SetReadDeadline(time.Time) error
}

var _ ReceiveGroupStream = (*receiveGroupStream)(nil)

type receiveGroupStream struct {
	subscribeID SubscribeID
	stream      transport.ReceiveStream

	sequence  GroupSequence
	startTime time.Time

	errCodeCh chan StreamErrorCode
}

func (r receiveGroupStream) SubscribeID() SubscribeID {
	return r.subscribeID
}

func (r receiveGroupStream) GroupSequence() GroupSequence {
	return r.sequence
}

func (r receiveGroupStream) ReadFrame() ([]byte, error) {
	var fm message.FrameMessage
	err := fm.Decode(r.stream)
	if err != nil {
		return nil, err
	}

	return fm.Payload, nil
}

func (r receiveGroupStream) StartAt() time.Time {
	return r.startTime
}

func (r receiveGroupStream) CancelRead(code StreamErrorCode) {
	if r.errCodeCh == nil {
		r.errCodeCh = make(chan StreamErrorCode, 1)
	}

	select {
	case r.errCodeCh <- code:
	default:
	}

	r.stream.CancelRead(transport.StreamErrorCode(code))
}

func (r receiveGroupStream) SetReadDeadline(t time.Time) error {
	return r.stream.SetReadDeadline(t)
}

func newGroupReceiverQueue() *groupReceiverQueue {
	return &groupReceiverQueue{
		queue: make([]ReceiveGroupStream, 0), // TODO: Tune the initial capacity
		ch:    make(chan struct{}, 1),
	}
}

type groupReceiverQueue struct {
	queue []ReceiveGroupStream
	ch    chan struct{}
	mu    sync.Mutex
}

func (q *groupReceiverQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *groupReceiverQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *groupReceiverQueue) Enqueue(stream ReceiveGroupStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, stream)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *groupReceiverQueue) Dequeue() ReceiveGroupStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	next := q.queue[0]

	q.queue = q.queue[1:]

	return next
}
