package moqt

import (
	"io"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

/*
 * Sequence number of a group in a track
 * When this is integer more than 1, the number means the sequence number.
 * When this is 0, it indicates the sequence number is currently unknown .
 * 0 is used to specify "the latest sequence number" or "the final sequence number of an open-ended track", "the first sequence number of the default order".
 */
type GroupSequence message.GroupSequence

/***/
type FetchRequest struct {
	SubscribeID   SubscribeID
	TrackPath     []string
	TrackPriority TrackPriority
	GroupSequence GroupSequence
	FrameSequence FrameSequence
}

func readFetch(r io.Reader) (FetchRequest, error) {
	var fm message.FetchMessage
	err := fm.Decode(r)
	if err != nil {
		slog.Error("failed to read a FETCH message", slog.String("error", err.Error()))
		return FetchRequest{}, err
	}

	req := FetchRequest{
		SubscribeID:   SubscribeID(fm.SubscribeID),
		TrackPath:     fm.TrackPath,
		TrackPriority: TrackPriority(fm.TrackPriority),
		GroupSequence: GroupSequence(fm.GroupSequence),
		FrameSequence: FrameSequence(fm.FrameSequence),
	}

	return req, nil
}

func writeFetch(w io.Writer, fetch FetchRequest) error {
	fm := message.FetchMessage{
		SubscribeID:   message.SubscribeID(fetch.SubscribeID),
		TrackPath:     fetch.TrackPath,
		TrackPriority: message.TrackPriority(fetch.TrackPriority),
		GroupSequence: message.GroupSequence(fetch.GroupSequence),
		FrameSequence: message.FrameSequence(fetch.FrameSequence),
	}
	err := fm.Encode(w)
	if err != nil {
		slog.Error("failed to send a FETCH message", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func newReceivedFetchQueue() *receiveFetchStreamQueue {
	return &receiveFetchStreamQueue{
		queue: make([]*receiveFetchStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type receiveFetchStreamQueue struct {
	queue []*receiveFetchStream
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *receiveFetchStreamQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *receiveFetchStreamQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *receiveFetchStreamQueue) Enqueue(fetch *receiveFetchStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, fetch)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receiveFetchStreamQueue) Dequeue() *receiveFetchStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	next := q.queue[0]
	q.queue = q.queue[1:]

	return next
}
