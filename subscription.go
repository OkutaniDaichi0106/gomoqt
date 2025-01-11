package moqt

import (
	"log/slog"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type SubscribeID uint64

type Subscription struct {
	/*
	 * Required
	 */
	TrackPath string

	/*
	 * Optional
	 */
	TrackPriority TrackPriority
	GroupOrder    GroupOrder
	GroupExpires  time.Duration

	// Parameters
	AuthorizationInfo string

	DeliveryTimeout time.Duration //TODO

	// AnnounceParameters Parameters

	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence

	SubscribeParameters Parameters
}

func newReceiveSubscribeStreamQueue() *receiveSubscribeStreamQueue {
	return &receiveSubscribeStreamQueue{
		queue: make([]ReceiveSubscribeStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type receiveSubscribeStreamQueue struct {
	queue []ReceiveSubscribeStream
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

func (q *receiveSubscribeStreamQueue) Enqueue(rss ReceiveSubscribeStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, rss)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receiveSubscribeStreamQueue) Dequeue() ReceiveSubscribeStream {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	next := q.queue[0]
	q.queue = q.queue[1:]

	return next
}

func readSubscription(r transport.Stream) (SubscribeID, Subscription, error) {
	var sm message.SubscribeMessage
	err := sm.Decode(r)
	if err != nil {
		slog.Debug("failed to read a SUBSCRIBE message", slog.String("error", err.Error()))
		return 0, Subscription{}, err
	}

	subscription := Subscription{
		TrackPath:           sm.TrackPath,
		TrackPriority:       TrackPriority(sm.TrackPriority),
		GroupOrder:          GroupOrder(sm.GroupOrder),
		GroupExpires:        sm.GroupExpires,
		MinGroupSequence:    GroupSequence(sm.MinGroupSequence),
		MaxGroupSequence:    GroupSequence(sm.MaxGroupSequence),
		SubscribeParameters: Parameters(sm.Parameters),
	}

	// Get a DELIVERY_TIMEOUT parameter
	deliveryTimeout, ok := getDeliveryTimeout(Parameters(sm.Parameters))
	if ok {
		subscription.DeliveryTimeout = deliveryTimeout
	}

	return SubscribeID(sm.SubscribeID), subscription, nil
}

func writeSubscription(w transport.Stream, id SubscribeID, subscription Subscription) error {
	// Set parameters
	if subscription.SubscribeParameters == nil {
		subscription.SubscribeParameters = make(Parameters)
	}

	// Set a DELIVERY_TIMEOUT parameter
	if subscription.DeliveryTimeout > 0 {
		subscription.SubscribeParameters.Add(DELIVERY_TIMEOUT, subscription.DeliveryTimeout)
	}

	// Send a SUBSCRIBE message
	sm := message.SubscribeMessage{
		SubscribeID:      message.SubscribeID(id),
		TrackPath:        subscription.TrackPath,
		TrackPriority:    message.TrackPriority(subscription.TrackPriority),
		GroupOrder:       message.GroupOrder(subscription.GroupOrder),
		GroupExpires:     subscription.GroupExpires,
		MinGroupSequence: message.GroupSequence(subscription.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(subscription.MaxGroupSequence),
		Parameters:       message.Parameters(subscription.SubscribeParameters),
	}
	err := sm.Encode(w)
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE message", slog.String("error", err.Error()))
		return err
	}

	return nil
}
