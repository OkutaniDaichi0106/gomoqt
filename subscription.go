package moqt

import (
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
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

type SendSubscribeStream struct {
	subscribeID SubscribeID
	Subscription
	stream transport.Stream
	mu     sync.Mutex
}

func (ss *SendSubscribeStream) SubscribeID() SubscribeID {
	return ss.subscribeID
}

func (ss *SendSubscribeStream) UpdateSubscribe(update SubscribeUpdate) error {
	/*
	 * Verify the update
	 */
	// Verify if the new group range is valid
	if update.MinGroupSequence > update.MaxGroupSequence {
		slog.Debug("MinGroupSequence is larger than MaxGroupSequence")
		return ErrInvalidRange
	}
	// Verify if the minimum group sequence become larger
	if ss.MinGroupSequence > update.MinGroupSequence {
		slog.Debug("the new MinGroupSequence is smaller than the old MinGroupSequence")
		return ErrInvalidRange
	}
	// Verify if the maximum group sequence become smaller
	if ss.MaxGroupSequence < update.MaxGroupSequence {
		slog.Debug("the new MaxGroupSequence is larger than the old MaxGroupSequence")
		return ErrInvalidRange
	}

	/*
	 * Send a SUBSCRIBE_UPDATE message
	 */
	// Set parameters
	if update.SubscribeParameters == nil {
		update.SubscribeParameters = make(Parameters)
	}
	if update.DeliveryTimeout > 0 {
		update.SubscribeParameters.Add(DELIVERY_TIMEOUT, update.DeliveryTimeout)
	}
	// Send a SUBSCRIBE_UPDATE message
	sum := message.SubscribeUpdateMessage{
		SubscribeID:      message.SubscribeID(ss.SubscribeID()),
		TrackPriority:    message.TrackPriority(update.TrackPriority),
		GroupOrder:       message.GroupOrder(update.GroupOrder),
		GroupExpires:     update.GroupExpires,
		MinGroupSequence: message.GroupSequence(update.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(update.MaxGroupSequence),
		Parameters:       message.Parameters(update.SubscribeParameters),
	}
	err := sum.Encode(ss.stream)
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
		return err
	}

	// Receive an INFO message
	info, err := readInfo(ss.stream)
	if err != nil {
		slog.Debug("failed to get an Info")
		return err
	}

	// Update the TrackPriority
	if info.TrackPriority == update.TrackPriority {
		ss.Track.TrackPriority = info.TrackPriority
	} else {
		slog.Debug("TrackPriority is not updated")
		return ErrPriorityMismatch
	}

	// Update the GroupOrder
	if update.GroupOrder == 0 {
		ss.Track.GroupOrder = info.GroupOrder
	} else {
		if info.GroupOrder != update.GroupOrder {
			slog.Debug("GroupOrder is not updated")
			return ErrGroupOrderMismatch
		}

		ss.Track.GroupOrder = update.GroupOrder
	}

	// Update the GroupExpires
	if info.GroupExpires < update.GroupExpires {
		ss.Track.GroupExpires = info.GroupExpires
	} else {
		ss.Track.GroupExpires = update.GroupExpires
	}

	// Update the MinGroupSequence and MaxGroupSequence
	ss.MinGroupSequence = update.MinGroupSequence
	ss.MaxGroupSequence = update.MaxGroupSequence

	// Update the SubscribeParameters
	ss.SubscribeParameters = update.SubscribeParameters

	// Update the DeliveryTimeout
	if update.DeliveryTimeout != 0 {
		ss.Track.DeliveryTimeout = update.DeliveryTimeout
	}

	return nil
}

func newReceivedSubscription(stream transport.Stream) (*ReceivedSubscribeStream, error) {
	id, subscription, err := readSubscription(stream)
	if err != nil {
		slog.Error("failed to get a subscription", slog.String("error", err.Error()))
		return nil, err
	}

	rs := &ReceivedSubscribeStream{
		subscribeID:  id,
		Subscription: subscription,
		stream:       stream,
	}

	// go rs.listenUpdate()

	return rs, nil
}

type ReceivedSubscribeStream struct {
	subscribeID SubscribeID
	Subscription
	stream transport.Stream
	mu     sync.Mutex
}

func (rss *ReceivedSubscribeStream) SubscribeID() SubscribeID {
	return rss.subscribeID
}

func (rss *ReceivedSubscribeStream) updateLastestGroupSequence(sequence GroupSequence) {
	atomic.StoreUint64((*uint64)(&rss.latestGroupSequence), uint64(sequence))
}

// func (rs *ReceivedSubscribeStream) listenUpdate() {
// 	for {
// 		err := func() error {
// 			// Read a SUBSCRIBE_UPDATE message
// 			update, err := readSubscribeUpdate(rs.stream)
// 			if err != nil {
// 				slog.Error("failed to receive an update", slog.String("error", err.Error()))
// 				return err
// 			}
// 			/*
// 			 * Update the subscription
// 			 */
// 			rs.mu.Lock()
// 			defer rs.mu.Unlock()

// 			// Verify the new range
// 			if update.MinGroupSequence > update.MaxGroupSequence {
// 				return ErrInvalidRange
// 			}

// 			// Update the track priority
// 			if update.TrackPriority != 0 {
// 				rs.Track.TrackPriority = update.TrackPriority
// 			}

// 			// Update the group expires
// 			if update.GroupExpires != 0 {
// 				rs.Track.GroupExpires = update.GroupExpires
// 			}

// 			// Update the group order
// 			if update.GroupOrder != 0 {
// 				rs.Track.GroupOrder = update.GroupOrder
// 			}

// 			// Update the group sequence range
// 			if update.MinGroupSequence != 0 {
// 				if rs.Subscription.MinGroupSequence > update.MinGroupSequence {
// 					return ErrInvalidRange
// 				}
// 				rs.Subscription.MinGroupSequence = update.MinGroupSequence
// 			}

// 			if update.MaxGroupSequence != 0 {
// 				if rs.Subscription.MaxGroupSequence < update.MaxGroupSequence {
// 					return ErrInvalidRange
// 				}
// 				rs.Subscription.MaxGroupSequence = update.MaxGroupSequence
// 			}

// 			rs.SubscribeParameters = update.SubscribeParameters

// 			if update.DeliveryTimeout != 0 {
// 				rs.Track.DeliveryTimeout = update.DeliveryTimeout
// 			}

// 			return nil
// 		}()

// 		if err != nil {
// 			slog.Error("failed to update the subscription", slog.String("error", err.Error()))
// 			rs.CloseWithError(err)
// 			return
// 		}
// 	}
// }

func (sess *session) OpenDataStream(subscription *ReceivedSubscribeStream, sequence GroupSequence, priority GroupPriority) (DataSendStream, error) {
	// Verify
	if sequence == 0 {
		return nil, errors.New("0 sequence number")
	}

	// Open
	stream, err := openGroupStream(sess.conn)
	if err != nil {
		slog.Error("failed to open a group stream", slog.String("error", err.Error()))
		return nil, err
	}

	// Send the GROUP message
	gm := message.GroupMessage{
		SubscribeID:   message.SubscribeID(subscription.SubscribeID()),
		GroupSequence: message.GroupSequence(sequence),
		GroupPriority: message.GroupPriority(priority),
	}
	err = gm.Encode(stream)
	if err != nil {
		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
		return nil, err
	}

	return dataSendStream{
			SendStream: stream,
			sentGroup: sentGroup{
				subscribeID:   subscription.SubscribeID(),
				groupSequence: sequence,
				groupPriority: priority,
				sentAt:        time.Now(),
			},
		},
		nil
}

func (sess *session) SendDatagram(subscription *ReceivedSubscribeStream, sequence GroupSequence, priority GroupPriority, payload []byte) (SentDatagram, error) {
	// Verify
	if sequence == 0 {
		return nil, errors.New("0 sequence number")

	}

	group := sentGroup{
		subscribeID:   subscription.SubscribeID(),
		groupSequence: sequence,
		groupPriority: priority,
		sentAt:        time.Now(),
	}

	// Send
	err := sendDatagram(sess.conn, group, payload)
	if err != nil {
		slog.Error("failed to send a datagram", slog.String("error", err.Error()))
		return nil, err
	}

	return &sentDatagram{
		payload:   payload,
		sentGroup: group,
	}, nil
}

func (rs *ReceivedSubscribeStream) CountDataGap(code uint64) error {
	// TODO: Implement
	sgm := message.SubscribeGapMessage{
		// GroupStartSequence: ,
		// Count: ,
		// GroupErrorCode: ,
	}
	err := sgm.Encode(rs.stream)
	if err != nil {
		slog.Error("failed to encode SUBSCRIBE_GAP message")
		return err
	}

	return nil
}

func (srs *ReceivedSubscribeStream) CloseWithError(err error) error {
	slog.Debug("closing a subscrbe receive stream", slog.Any("subscription", srs.Subscription))

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

	slog.Debug("closed a subscrbe receive stream", slog.Any("subscription", srs.Subscription))

	return nil
}

func (srs *ReceivedSubscribeStream) Close() error {
	slog.Debug("closing a subscrbe receive stream", slog.Any("subscription", srs.Subscription))

	err := srs.stream.Close()
	if err != nil {
		slog.Debug("catch an error when closing a Subscribe Stream", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("closed a subscrbe receive stream", slog.Any("subscription", srs.Subscription))

	return nil
}

func newReceivedSubscriptionQueue() *receivedSubscriptionQueue {
	return &receivedSubscriptionQueue{
		queue: make([]*ReceivedSubscribeStream, 0),
		ch:    make(chan struct{}, 1),
	}
}

type receivedSubscriptionQueue struct {
	queue []*ReceivedSubscribeStream
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

func (q *receivedSubscriptionQueue) Enqueue(subscription *ReceivedSubscribeStream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = append(q.queue, subscription)

	select {
	case q.ch <- struct{}{}:
	default:
	}
}

func (q *receivedSubscriptionQueue) Dequeue() *ReceivedSubscribeStream {
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
