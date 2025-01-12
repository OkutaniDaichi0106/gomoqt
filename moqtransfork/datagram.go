package moqtransfork

import (
	"bytes"
	"sync"
)

/*
 *
 *
 */
func newReceivedDatagram(datagram []byte) (ReceivedDatagram, error) {
	//
	reader := bytes.NewReader(datagram)

	//
	id, group, err := readGroup(reader)
	if err != nil {
		return nil, err
	}

	//
	frame := datagram[len(datagram)-reader.Len():]

	return &receivedDatagram{
		subscribeID:   id,
		ReceivedGroup: group,
		payload:       frame,
	}, nil
}

type ReceivedDatagram interface {
	SubscribeID() SubscribeID
	Payload() []byte
	ReceivedGroup
}

var _ ReceivedDatagram = (*receivedDatagram)(nil)

type receivedDatagram struct {
	subscribeID SubscribeID
	ReceivedGroup
	payload []byte
}

func (d receivedDatagram) SubscribeID() SubscribeID {
	return d.subscribeID
}

func (d receivedDatagram) Payload() []byte {
	return d.payload
}

/*
 *
 *
 */
func newReceivedDatagramQueue() *receivedDatagramQueue {
	return &receivedDatagramQueue{
		queue: make([]ReceivedDatagram, 0), // Tune the initial capacity
		ch:    make(chan struct{}, 1),
	}
}

type receivedDatagramQueue struct {
	mu    sync.Mutex
	queue []ReceivedDatagram
	ch    chan struct{}
}

func (q *receivedDatagramQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *receivedDatagramQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *receivedDatagramQueue) Enqueue(datagram ReceivedDatagram) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, datagram)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receivedDatagramQueue) Dequeue() ReceivedDatagram {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	next := q.queue[0]
	q.queue = q.queue[1:]

	return next
}

type SentDatagram interface {
	Payload() []byte
	Group
}

var _ SentDatagram = (*sentDatagram)(nil)

type sentDatagram struct {
	sentGroup
	payload []byte
}

func (d sentDatagram) Payload() []byte {
	return d.payload
}

func (d *sentDatagram) Write(buf []byte) (int, error) {
	return copy(d.payload, buf), nil
}

// type sentDatagramQueue struct {
// 	mu    sync.Mutex
// 	queue []SentDatagram
// 	ch    chan struct{}
// }

// func (q *sentDatagramQueue) Len() int {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	return len(q.queue)
// }

// func (q *sentDatagramQueue) Chan() <-chan struct{} {

// 	return q.ch
// }

// func (q *sentDatagramQueue) Enqueue(datagram SentDatagram) {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	q.queue = append(q.queue, datagram)

// 	select {
// 	case q.ch <- struct{}{}:
// 	default:
// 	}
// }

// func (q *sentDatagramQueue) Dequeue() SentDatagram {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()

// 	if len(q.queue) == 0 {
// 		return nil
// 	}

// 	next := q.queue[0]
// 	q.queue = q.queue[1:]

// 	return next
// }
