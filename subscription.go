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

type SubscribeID uint64

type Subscription struct {
	//subscribeID SubscribeID

	Track Track

	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence

	SubscribeParameters Parameters

	/*
	 * Not in wire
	 */
}

type SentSubscription struct {
	subscribeID SubscribeID
	Subscription
	stream transport.Stream
	mu     sync.Mutex
}

func (ss *SentSubscription) SubscribeID() SubscribeID {
	return ss.subscribeID
}

func (ss *SentSubscription) TrackInfo() Info {
	return ss.Track.Info()
}

func newReceivedSubscription(stream transport.Stream) (*ReceivedSubscription, error) {
	id, subscription, err := readSubscription(stream)
	if err != nil {
		slog.Error("failed to get a subscription", slog.String("error", err.Error()))
		return nil, err
	}

	return &ReceivedSubscription{
		subscribeID:  id,
		Subscription: subscription,
		stream:       stream,
	}, nil
}

type ReceivedSubscription struct {
	subscribeID SubscribeID
	Subscription
	stream transport.Stream
	mu     sync.Mutex
}

func (rs *ReceivedSubscription) SubscribeID() SubscribeID {
	return rs.subscribeID
}

func (srs *ReceivedSubscription) Inform(info Info) error {
	slog.Debug("Accepting the subscription")

	im := message.InfoMessage{
		GroupPriority:       message.GroupPriority(srs.Track.TrackPriority),
		LatestGroupSequence: message.GroupSequence(info.LatestGroupSequence),
		GroupOrder:          message.GroupOrder(info.GroupOrder),
		GroupExpires:        info.GroupExpires,
	}

	err := im.Encode(srs.stream)
	if err != nil {
		slog.Error("failed to inform track status", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Informed", slog.Any("info", info))

	return nil
}

func (srs *ReceivedSubscription) ReceiveUpdate() (SubscribeUpdate, error) {

	return readSubscribeUpdate(srs.stream)
}

func (rs *ReceivedSubscription) UpdateSubscription(update SubscribeUpdate) error {
	// Verify
	if update.MinGroupSequence > update.MaxGroupSequence {
		return ErrInvalidRange
	}

	// Update the subscription
	if update.TrackPriority != 0 {
		rs.Track.TrackPriority = update.TrackPriority
	}

	if update.GroupExpires != 0 {
		rs.Track.GroupExpires = update.GroupExpires
	}

	if update.GroupOrder != 0 {
		rs.Track.GroupOrder = update.GroupOrder
	}

	if update.MinGroupSequence != 0 {
		if rs.Subscription.MinGroupSequence > update.MinGroupSequence {
			return ErrInvalidRange
		}
		rs.Subscription.MinGroupSequence = update.MinGroupSequence
	}

	if update.MaxGroupSequence != 0 {
		if rs.Subscription.MaxGroupSequence < update.MaxGroupSequence {
			return ErrInvalidRange
		}
		rs.Subscription.MaxGroupSequence = update.MaxGroupSequence
	}

	rs.SubscribeParameters = update.SubscribeParameters

	if update.DeliveryTimeout != 0 {
		rs.Track.DeliveryTimeout = update.DeliveryTimeout
	}

	return nil
}

func (rs *ReceivedSubscription) CountDataGap(err error) error {
	// TODO
	sgm := message.SubscribeGapMessage{}
	err = sgm.Encode(rs.stream)
	if err != nil {
		slog.Error("failed to encode SUBSCRIBE_GAP message")
		return err
	}

	return nil
}

func (srs *ReceivedSubscription) CloseWithError(err error) {
	slog.Debug("closing a subscrbe receive stream", slog.Any("subscription", srs.Subscription))

	if err == nil {
		srs.Close()
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

	slog.Debug("closed a subscrbe receive stream", slog.Any("subscription", srs.Subscription))
}

func (srs *ReceivedSubscription) Close() {
	slog.Debug("closing a subscrbe receive stream", slog.Any("subscription", srs.Subscription))

	err := srs.stream.Close()
	if err != nil {
		slog.Debug("catch an error when closing a Subscribe Stream", slog.String("error", err.Error()))
	}

	slog.Debug("closed a subscrbe receive stream", slog.Any("subscription", srs.Subscription))
}

type receivedSubscriptionQueue struct {
	queue []*ReceivedSubscription
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

func (q *receivedSubscriptionQueue) Enqueue(subscription *ReceivedSubscription) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, subscription)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receivedSubscriptionQueue) Dequeue() *ReceivedSubscription {
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
