package moqt

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type SendInfoStream interface {
	InfoRequest() InfoRequest
	SendInfo(Info) error
	CloseWithError(error) error
	Close() error
}

var _ SendInfoStream = (*sendInfoStream)(nil)

type sendInfoStream struct {
	req    InfoRequest
	stream transport.Stream
	mu     sync.Mutex

	// Used to signal that a subscribe stream is
	ch chan struct{}
}

func (req *sendInfoStream) InfoRequest() InfoRequest {
	return req.req
}

func (req *sendInfoStream) SendInfo(i Info) error {
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

	slog.Info("answered an info")

	if req.ch != nil {
		close(req.ch)
		req.ch = nil
	}

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

func newReceiveInfoStreamQueue() *receiveInfoStreamQueue {
	return &receiveInfoStreamQueue{
		queue: make([]*sendInfoStream, 0),
		ch:    make(chan struct{}),
	}
}

type receiveInfoStreamQueue struct {
	queue []*sendInfoStream
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *receiveInfoStreamQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *receiveInfoStreamQueue) Enqueue(req *sendInfoStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, req)
}

func (q *receiveInfoStreamQueue) Dequeue() *sendInfoStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	req := q.queue[0]
	q.queue = q.queue[1:]

	return req
}

func (q *receiveInfoStreamQueue) Chan() <-chan struct{} {
	return q.ch
}
