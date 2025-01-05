package moqt

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

/*
 * Sequence number of a group in a track
 * When this is integer more than 1, the number means the sequence number.
 * When this is 0, it indicates the sequence number is currently unknown .
 * 0 is used to specify "the latest sequence number" or "the final sequence number of an open-ended track", "the first sequence number of the default order".
 */
type GroupSequence message.GroupSequence

/***/
type Fetch struct {
	TrackPath     string
	GroupPriority GroupPriority
	GroupSequence GroupSequence
	FrameSequence FrameSequence
}

func newReceivedFetch(stream transport.Stream) (*receiveFetchStream, error) {
	// Get a fetch-request
	fetch, err := readFetch(stream)
	if err != nil {
		slog.Error("failed to get a fetch-request", slog.String("error", err.Error()))
		return nil, err
	}

	return &receiveFetchStream{
		fetch:  fetch,
		stream: stream,
	}, nil
}

type ReceiveFetchStream interface {
	OpenDataStream(SubscribeID, GroupSequence, GroupPriority) (SendDataStream, error)
	CloseWithError(error) error
	Close() error
}

var _ ReceiveFetchStream = (*receiveFetchStream)(nil)

type receiveFetchStream struct {
	fetch     Fetch
	groupSent bool
	stream    transport.Stream
}

func (fetch *receiveFetchStream) OpenDataStream(id SubscribeID, sequence GroupSequence, priority GroupPriority) (SendDataStream, error) {
	if fetch.groupSent {
		return nil, errors.New("a group has already been sent")
	}

	// Send a GROUP message
	gm := message.GroupMessage{
		SubscribeID:   message.SubscribeID(id),
		GroupSequence: message.GroupSequence(sequence),
		GroupPriority: message.GroupPriority(priority),
	}
	err := gm.Encode(fetch.stream)
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		return nil, err
	}

	fetch.groupSent = true

	return dataSendStream{
		SendStream: fetch.stream,
		sentGroup: sentGroup{
			subscribeID:   id,
			groupSequence: sequence,
			groupPriority: priority,
			sentAt:        time.Now(),
		},
	}, nil
}

func (fetch receiveFetchStream) CloseWithError(err error) error {
	if err == nil {
		return fetch.Close()
	}

	var code transport.StreamErrorCode

	var strerr transport.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		feterr, ok := err.(FetchError)
		if ok {
			code = transport.StreamErrorCode(feterr.FetchErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	fetch.stream.CancelRead(code)
	fetch.stream.CancelWrite(code)

	slog.Info("rejcted the fetch request")

	return nil
}

func (frw receiveFetchStream) Close() error {
	return frw.stream.Close()
}

func newReceivedFetchQueue() *receivedFetchQueue {
	return &receivedFetchQueue{
		queue: make([]*receiveFetchStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type receivedFetchQueue struct {
	queue []*receiveFetchStream
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *receivedFetchQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *receivedFetchQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *receivedFetchQueue) Enqueue(fetch *receiveFetchStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, fetch)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receivedFetchQueue) Dequeue() *receiveFetchStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	next := q.queue[0]
	q.queue = q.queue[1:]

	return next
}
