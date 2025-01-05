package moqt

import (
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

// type InfoHandler interface {
// 	HandleInfo(Info)
// }

type InfoRequestHandler interface {
	HandleInfoRequest(InfoRequest, *Info, sendInfoStream)
}

type InfoRequest struct {
	TrackPath string
}

type Info struct {
	TrackPriority       TrackPriority
	LatestGroupSequence GroupSequence
	GroupOrder          GroupOrder
	GroupExpires        time.Duration
}

func newReceivedInfoRequest(stream transport.Stream) (*sendInfoStream, error) {
	req, err := readInfoRequest(stream)
	if err != nil {
		slog.Error("failed to get a info-request", slog.String("error", err.Error()))
		return nil, err
	}

	return &sendInfoStream{
		InfoRequest: req,
		stream:      stream,
	}, nil
}

type SendInfoStream interface {
	UpdateInfo(Info)
	CloseWithError(error) error
	Close() error
}

var _ SendInfoStream = (*sendInfoStream)(nil)

type sendInfoStream struct {
	InfoRequest
	stream transport.Stream
	mu     sync.Mutex
}

func (req *sendInfoStream) UpdateInfo(i Info) {
	req.mu.Lock()
	defer req.mu.Unlock()

	im := message.InfoMessage{
		GroupPriority:       message.GroupPriority(i.TrackPriority),
		LatestGroupSequence: message.GroupSequence(i.LatestGroupSequence),
		GroupOrder:          message.GroupOrder(i.GroupOrder),
		GroupExpires:        i.GroupExpires,
	}

	err := im.Encode(req.stream)
	if err != nil {
		slog.Error("failed to send an INFO message", slog.String("error", err.Error()))
		req.CloseWithError(err)
		return
	}

	slog.Info("answered an info")

	req.Close()
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

func newReceivedInfoRequestQueue() *receivedInfoRequestQueue {
	return &receivedInfoRequestQueue{
		queue: make([]*sendInfoStream, 0),
		ch:    make(chan struct{}),
	}
}

type receivedInfoRequestQueue struct {
	queue []*sendInfoStream
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *receivedInfoRequestQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *receivedInfoRequestQueue) Enqueue(req *sendInfoStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, req)
}

func (q *receivedInfoRequestQueue) Dequeue() *sendInfoStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	req := q.queue[0]
	q.queue = q.queue[1:]

	return req
}

func (q *receivedInfoRequestQueue) Chan() <-chan struct{} {
	return q.ch
}

func readInfo(r io.Reader) (Info, error) {
	// Read an INFO message
	var im message.InfoMessage
	err := im.Decode(r)
	if err != nil {
		slog.Error("failed to read a INFO message", slog.String("error", err.Error()))
		return Info{}, err
	}

	info := Info{
		TrackPriority:       TrackPriority(im.GroupPriority),
		LatestGroupSequence: GroupSequence(im.LatestGroupSequence),
		GroupOrder:          GroupOrder(im.GroupOrder),
		GroupExpires:        im.GroupExpires,
	}

	return info, nil
}
