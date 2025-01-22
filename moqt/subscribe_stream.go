package moqt

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
)

type SendSubscribeStream interface {
	// Get the SubscribeID
	SubscribeID() SubscribeID

	// Get the subscription
	SubscribeConfig() SubscribeConfig

	// Update the subscription
	UpdateSubscribe(SubscribeUpdate) error

	//
	ReceiveSubscribeGap() (SubscribeGap, error)

	// Close the stream
	Close() error

	// Close the stream with an error
	CloseWithError(error) error
}

var _ SendSubscribeStream = (*sendSubscribeStream)(nil)

type sendSubscribeStream struct {
	subscribeID  SubscribeID
	subscription SubscribeConfig
	stream       transport.Stream
	mu           sync.Mutex
}

func (ss *sendSubscribeStream) SubscribeID() SubscribeID {
	return ss.subscribeID
}

func (ss *sendSubscribeStream) SubscribeConfig() SubscribeConfig {
	return ss.subscription
}

func (sss *sendSubscribeStream) UpdateSubscribe(update SubscribeUpdate) error {
	sss.mu.Lock()
	defer sss.mu.Unlock()

	subscription, err := updateSubscription(sss.subscription, update)
	if err != nil {
		slog.Error("failed to update a subscription", slog.String("error", err.Error()))
		return err
	}

	err = writeSubscribeUpdate(sss.stream, update)
	if err != nil {
		slog.Error("failed to write a subscribe update message", slog.String("error", err.Error()))
		return err
	}

	sss.subscription = subscription

	slog.Debug("updated a subscription", slog.Any("subscription", sss.subscription))

	return nil
}
func (ss *sendSubscribeStream) ReceiveSubscribeGap() (SubscribeGap, error) {
	slog.Debug("receiving a data gap")

	gap, err := readSubscribeGap(ss.stream)
	if err != nil {
		slog.Error("failed to read a subscribe gap message", slog.String("error", err.Error()))
		return SubscribeGap{}, err
	}

	slog.Debug("received a data gap", slog.Any("gap", gap))

	return gap, nil
}

func (ss *sendSubscribeStream) Close() error {
	slog.Debug("closing a subscrbe send stream", slog.Any("subscription", ss.subscription))

	err := ss.stream.Close()
	if err != nil {
		slog.Debug("catch an error when closing a Subscribe Stream", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("closed a subscrbe send stream", slog.Any("subscription", ss.subscription))

	return nil
}

func (sss *sendSubscribeStream) CloseWithError(err error) error {
	slog.Debug("closing a subscrbe send stream", slog.Any("subscription", sss.subscription))

	if err == nil {
		return sss.Close()
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

	sss.stream.CancelRead(code)
	sss.stream.CancelWrite(code)

	slog.Debug("closed a subscrbe receive stream", slog.Any("config", sss.SubscribeConfig()))

	return nil
}

type ReceiveSubscribeStream interface {
	SubscribeID() SubscribeID
	SubscribeConfig() SubscribeConfig

	SendSubscribeGap(SubscribeGap) error

	CloseWithError(error) error
	Close() error
}

var _ ReceiveSubscribeStream = (*receiveSubscribeStream)(nil)

type receiveSubscribeStream struct {
	subscribeID  SubscribeID
	subscription SubscribeConfig
	stream       transport.Stream
	mu           sync.Mutex
}

func (rss *receiveSubscribeStream) SubscribeID() SubscribeID {
	return rss.subscribeID
}

func (rss *receiveSubscribeStream) SubscribeConfig() SubscribeConfig {
	return rss.subscription
}

func (rss *receiveSubscribeStream) SendSubscribeGap(gap SubscribeGap) error {
	slog.Debug("sending a data gap", slog.Any("gap", gap))

	rss.mu.Lock()
	defer rss.mu.Unlock()

	err := writeSubscribeGap(rss.stream, gap)
	if err != nil {
		slog.Error("failed to write a subscribe gap message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("sent a data gap", slog.Any("gap", gap))

	return nil
}

func (srs *receiveSubscribeStream) CloseWithError(err error) error {
	slog.Debug("closing a subscrbe receive stream", slog.Any("subscription", srs.subscription))

	if err == nil {
		return srs.Close()
	}

	// TODO:

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

	srs.stream.CancelRead(code)
	srs.stream.CancelWrite(code)

	slog.Debug("closed a subscrbe receive stream", slog.Any("subscription", srs.subscription))

	return nil
}

func (srs *receiveSubscribeStream) Close() error {
	slog.Debug("closing a subscrbe receive stream", slog.Any("subscription", srs.subscription))

	err := srs.stream.Close()
	if err != nil {
		slog.Debug("catch an error when closing a Subscribe Stream", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("closed a subscrbe receive stream", slog.Any("subscription", srs.subscription))

	return nil
}

func newReceiveSubscribeStreamQueue() *receiveSubscribeStreamQueue {
	return &receiveSubscribeStreamQueue{
		queue: make([]*receiveSubscribeStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type receiveSubscribeStreamQueue struct {
	queue []*receiveSubscribeStream
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *receiveSubscribeStreamQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *receiveSubscribeStreamQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *receiveSubscribeStreamQueue) Enqueue(rss *receiveSubscribeStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, rss)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receiveSubscribeStreamQueue) Dequeue() *receiveSubscribeStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	next := q.queue[0]
	q.queue = q.queue[1:]

	return next
}
