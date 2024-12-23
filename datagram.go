package moqt

import (
	"bytes"
	"io"
	"log/slog"
	"sync"
)

type ReceivedDatagram interface {
	io.Reader
	Payload() []byte
	Group
}

func newReceivedDatagram(datagram []byte) (ReceivedDatagram, error) {
	// Get a payload reader
	reader := bytes.NewReader(datagram)

	// Read a group
	group, err := readGroup(reader)
	if err != nil {
		slog.Error("failed to get a group", slog.String("error", err.Error()))
		return nil, err
	}

	return &receivedDatagram{
		ReceivedGroup: group,
		payload:       datagram[len(datagram)-reader.Len():],
	}, nil
}

var _ ReceivedDatagram = (*receivedDatagram)(nil)

type receivedDatagram struct {
	ReceivedGroup
	payload []byte
}

func (d receivedDatagram) Payload() []byte {
	return d.payload
}

func (d receivedDatagram) Read(buf []byte) (int, error) {
	return copy(buf, d.payload), nil
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
	io.Writer
	Payload() []byte
	Group
}

func newSentDatagram(group SentGroup, payload []byte) SentDatagram {
	return &sentDatagram{
		SentGroup: group,
		payload:   payload,
	}
}

var _ SentDatagram = (*sentDatagram)(nil)

type sentDatagram struct {
	SentGroup
	payload []byte
}

func (d sentDatagram) Payload() []byte {
	return d.payload
}

func (d *sentDatagram) Write(buf []byte) (int, error) {
	return copy(d.payload, buf), nil
}

type sentDatagramQueue struct {
	mu    sync.Mutex
	queue []SentDatagram
	ch    chan struct{}
}

func (q *sentDatagramQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *sentDatagramQueue) Chan() <-chan struct{} {

	return q.ch
}

func (q *sentDatagramQueue) Enqueue(datagram SentDatagram) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, datagram)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *sentDatagramQueue) Dequeue() SentDatagram {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	next := q.queue[0]
	q.queue = q.queue[1:]

	return next
}
