package moqt

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

type sendInfoStream struct {
	req    InfoRequest
	stream transport.Stream
	mu     sync.Mutex
}

func (req *sendInfoStream) InfoRequest() InfoRequest {
	return req.req
}

func (req *sendInfoStream) SendInfoAndClose(i Info) error {
	req.mu.Lock()
	defer req.mu.Unlock()

	im := message.InfoMessage{
		TrackPriority:       message.TrackPriority(i.TrackPriority),
		LatestGroupSequence: message.GroupSequence(i.LatestGroupSequence),
		GroupOrder:          message.GroupOrder(i.GroupOrder),
	}

	err := im.Encode(req.stream)
	if err != nil {
		slog.Error("failed to send an INFO message", slog.String("error", err.Error()))
		return err
	}

	slog.Info("sended an info")

	req.Close()

	return nil
}

func (req *sendInfoStream) CloseWithError(err error) error {
	req.mu.Lock()
	defer req.mu.Unlock()

	if err == nil {
		return req.Close()
	}

	req.mu.Lock()
	defer req.mu.Unlock()

	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		inferr, ok := err.(InfoError)
		if ok {
			code = transport.StreamErrorCode(inferr.InfoErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	req.stream.CancelRead(code)
	req.stream.CancelWrite(code)

	slog.Info("rejected an info request")

	return nil
}

func (req *sendInfoStream) Close() error {
	req.mu.Lock()
	defer req.mu.Unlock()

	return req.stream.Close()
}

func newReceiveInfoStreamQueue() *sendInfoStreamQueue {
	return &sendInfoStreamQueue{
		queue: make([]*sendInfoStream, 0),
		ch:    make(chan struct{}),
	}
}

type sendInfoStreamQueue struct {
	queue []*sendInfoStream
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *sendInfoStreamQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *sendInfoStreamQueue) Enqueue(req *sendInfoStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, req)
}

func (q *sendInfoStreamQueue) Dequeue() *sendInfoStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	req := q.queue[0]
	q.queue = q.queue[1:]

	return req
}

func (q *sendInfoStreamQueue) Chan() <-chan struct{} {
	return q.ch
}
