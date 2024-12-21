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
 * Fetch Stream
 */

type FetchStream struct {
	stream transport.Stream
}

func (f FetchStream) Read(buf []byte) (int, error) {
	return f.stream.Read(buf)
}

// func (f FetchStream) Group() Group {
// 	return f.group
// }

func (f FetchStream) CancelRead(code transport.StreamErrorCode) {
	f.stream.CancelRead(code)
}

func (f FetchStream) Close() error {
	err := f.stream.Close()
	if err != nil {
		slog.Error("failed to close a Fetch Stream", slog.String("error", err.Error()))
		return err
	}

	return nil
}

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

func newReceivedFetch(stream transport.Stream) (*ReceivedFetch, error) {
	// Get a fetch-request
	fetch, err := readFetch(stream)
	if err != nil {
		slog.Error("failed to get a fetch-request", slog.String("error", err.Error()))
		return nil, err
	}

	return &ReceivedFetch{
		fetch:  fetch,
		stream: stream,
	}, nil
}

type ReceivedFetch struct {
	fetch     Fetch
	groupSent bool
	stream    transport.Stream
}

func (fetch *ReceivedFetch) OpenDataStream(id SubscribeID, sequence GroupSequence, priority GroupPriority) (DataSendStream, error) {
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
		SentGroup: SentGroup{
			subscribeID:   id,
			groupSequence: sequence,
			groupPriority: priority,
			sentAt:        time.Now(),
		},
	}, nil
}

func (fetch ReceivedFetch) Reject(err error) {
	if err == nil {
		fetch.Close()
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
}

func (frw ReceivedFetch) Close() error {
	return frw.stream.Close()
}

type receivedFetchQueue struct {
	queue []*ReceivedFetch
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

func (q *receivedFetchQueue) Enqueue(fetch *ReceivedFetch) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, fetch)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receivedFetchQueue) Dequeue() *ReceivedFetch {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	next := q.queue[0]
	q.queue = q.queue[1:]

	return next
}
