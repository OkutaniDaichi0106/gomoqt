package moqt

import (
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type SubscribeID uint64

type Subscription struct {
	Track

	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence

	SubscribeParameters Parameters

	/*
	 * Not in wire
	 */
}

func newReceivedSubscriptionQueue() *receivedSubscriptionQueue {
	return &receivedSubscriptionQueue{
		queue: make([]*receiveSubscribeStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type receivedSubscriptionQueue struct {
	queue []*receiveSubscribeStream
	mu    sync.Mutex
	ch    chan struct{}
}

func (q *receivedSubscriptionQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.queue)
}

func (q *receivedSubscriptionQueue) Chan() <-chan struct{} {
	return q.ch
}

func (q *receivedSubscriptionQueue) Enqueue(subscription *receiveSubscribeStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, subscription)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receivedSubscriptionQueue) Dequeue() *receiveSubscribeStream {
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
		Track: Track{
			TrackPath:     sm.TrackPath,
			TrackPriority: TrackPriority(sm.TrackPriority),
			GroupOrder:    GroupOrder(sm.GroupOrder),
			GroupExpires:  sm.GroupExpires,
		},
		// TODO: Delivery Timeout
		MinGroupSequence:    GroupSequence(sm.MinGroupSequence),
		MaxGroupSequence:    GroupSequence(sm.MaxGroupSequence),
		SubscribeParameters: Parameters(sm.Parameters),
	}

	// Get
	deliveryTimeout, ok := getDeliveryTimeout(Parameters(sm.Parameters))
	if ok {
		subscription.Track.DeliveryTimeout = deliveryTimeout
	}

	return SubscribeID(sm.SubscribeID), subscription, nil
}

type SubscribeUpdate struct {
	TrackPriority    TrackPriority
	GroupOrder       GroupOrder
	GroupExpires     time.Duration
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence

	/*
	 * SubscribeParameters
	 */
	SubscribeParameters Parameters

	DeliveryTimeout time.Duration
}

func readSubscribeUpdate(r io.Reader) (SubscribeUpdate, error) {

	// Read a SUBSCRIBE_UPDATE message
	var sum message.SubscribeUpdateMessage
	err := sum.Decode(r)
	if err != nil {
		slog.Debug("failed to read a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
		return SubscribeUpdate{}, err
	}

	// Get a DELIVERY_TIMEOUT parameter
	timeout, ok := getDeliveryTimeout(Parameters(sum.Parameters))
	if !ok {
		timeout = 0
	}

	return SubscribeUpdate{
		TrackPriority:       TrackPriority(sum.TrackPriority),
		GroupOrder:          GroupOrder(sum.GroupOrder),
		GroupExpires:        sum.GroupExpires,
		MinGroupSequence:    GroupSequence(sum.MinGroupSequence),
		MaxGroupSequence:    GroupSequence(sum.MaxGroupSequence),
		SubscribeParameters: Parameters(sum.Parameters),
		DeliveryTimeout:     timeout,
	}, nil
}
