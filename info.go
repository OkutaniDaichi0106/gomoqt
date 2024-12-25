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
	HandleInfoRequest(InfoRequest, *Info, ReceivedInfoRequest)
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

func newReceivedInfoRequest(stream transport.Stream) (*ReceivedInfoRequest, error) {
	req, err := readInfoRequest(stream)
	if err != nil {
		slog.Error("failed to get a info-request", slog.String("error", err.Error()))
		return nil, err
	}

	return &ReceivedInfoRequest{
		InfoRequest: req,
		stream:      stream,
	}, nil
}

type ReceivedInfoRequest struct {
	InfoRequest
	stream transport.Stream
	mu     sync.Mutex
}

func (req *ReceivedInfoRequest) Inform(i Info) {
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

func (req *ReceivedInfoRequest) CloseWithError(err error) error {
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

func (req *ReceivedInfoRequest) Close() error {
	req.mu.Lock()
	defer req.mu.Unlock()

	return req.stream.Close()
}

func newReceivedInfoRequestQueue() *receivedInfoRequestQueue {
	return &receivedInfoRequestQueue{
		queue: make([]*ReceivedInfoRequest, 0),
		ch:    make(chan struct{}),
	}
}

type receivedInfoRequestQueue struct {
	queue []*ReceivedInfoRequest
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *receivedInfoRequestQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *receivedInfoRequestQueue) Enqueue(req *ReceivedInfoRequest) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, req)
}

func (q *receivedInfoRequestQueue) Dequeue() *ReceivedInfoRequest {
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
