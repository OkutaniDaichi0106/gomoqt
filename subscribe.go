package moqt

import (
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type SubscribeID uint64

type Subscription struct {
	subscribeID        SubscribeID
	TrackPath          string
	SubscriberPriority SubscriberPriority
	GroupOrder         GroupOrder
	GroupExpires       time.Duration
	MinGroupSequence   GroupSequence
	MaxGroupSequence   GroupSequence

	/*
	 * Parameters
	 */
	Parameters Parameters

	DeliveryTimeout time.Duration //TODO
}

func (s Subscription) FirstGrouopSequence() GroupSequence {
	switch s.GroupOrder {
	case ASCENDING, DEFAULT:
		return s.MinGroupSequence
	case DESCENDING:
		return s.MaxGroupSequence
	default:
		return 0
	}
}

func (s Subscription) GetGroup(seq GroupSequence, priority PublisherPriority) Group {
	return Group{
		subscribeID:       s.subscribeID,
		groupSequence:     seq,
		PublisherPriority: priority,
	}
}

type SubscribeSender struct {
	stream       moq.Stream
	subscription Subscription
}

func (w *SubscribeSender) Update(update SubscribeUpdate) (Info, error) {
	old := w.subscription
	slog.Debug("updating a subscription", slog.Any("subscription", w.subscription), slog.Any("to", update))

	// Verify if the new group range is valid
	if update.MinGroupSequence > update.MaxGroupSequence {
		slog.Debug("MinGroupSequence is larger than MaxGroupSequence")
		return Info{}, ErrInvalidRange
	}
	//
	if old.MinGroupSequence > update.MinGroupSequence {
		slog.Debug("the new MinGroupSequence is smaller than the old MinGroupSequence")
		return Info{}, ErrInvalidRange
	}
	//
	if old.MaxGroupSequence < update.MaxGroupSequence {
		slog.Debug("the new MaxGroupSequence is larger than the old MaxGroupSequence")
		return Info{}, ErrInvalidRange
	}

	/*
	 * Send a SUBSCRIBE_UPDATE message
	 */
	// Set parameters
	if update.Parameters == nil {
		update.Parameters = make(Parameters)
	}
	if update.DeliveryTimeout > 0 {
		update.Parameters.Add(DELIVERY_TIMEOUT, update.DeliveryTimeout)
	}
	// Initialize
	sum := message.SubscribeUpdateMessage{
		SubscribeID:        message.SubscribeID(w.subscription.subscribeID),
		SubscriberPriority: message.SubscriberPriority(update.SubscriberPriority),
		GroupOrder:         message.GroupOrder(update.GroupOrder),
		GroupExpires:       update.GroupExpires,
		MinGroupSequence:   message.GroupSequence(update.MinGroupSequence),
		MaxGroupSequence:   message.GroupSequence(update.MaxGroupSequence),
		Parameters:         message.Parameters(update.Parameters),
	}

	err := sum.Encode(w.stream)
	if err != nil {
		slog.Debug("failed to send a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
		return Info{}, err
	}

	info, err := readInfo(w.stream)
	if err != nil {
		slog.Debug("failed to get an Info")
		return Info{}, err
	}

	return info, nil
}

func (s *SubscribeSender) Unsubscribe(err error) {
	slog.Debug("stopping a subscription", slog.String("reason", err.Error()))

	if err == nil {
		s.Close()
	}

	suberr, ok := err.(SubscribeError)
	if !ok {
		suberr = ErrInternalError
	}

	s.stream.CancelWrite(moq.StreamErrorCode(suberr.SubscribeErrorCode()))
	s.stream.CancelRead(moq.StreamErrorCode(suberr.SubscribeErrorCode()))
}

func (s SubscribeSender) Close() {
	err := s.stream.Close()
	if err != nil {
		slog.Error("catch an error when closing a subscribe stream", slog.String("error", err.Error()))
	}
}

func (s SubscribeSender) Subscription() Subscription {
	return s.subscription
}

/*
 *
 */

type SubscribeReceiver struct {
	subscription Subscription
	stream       moq.Stream
}

func (sr SubscribeReceiver) Subscription() Subscription {
	return sr.subscription
}

func (sr *SubscribeReceiver) updateSubscription(update SubscribeUpdate) {
	// Update the subscriber priority
	if update.SubscriberPriority != 0 {
		sr.subscription.SubscriberPriority = update.SubscriberPriority
	}

	// Update the group order
	if update.GroupOrder != 0 {
		sr.subscription.GroupOrder = update.GroupOrder
	}

	// Update the group expires
	if update.GroupExpires != 0 {
		sr.subscription.GroupExpires = update.GroupExpires
	}

	// Update the min group sequence
	if update.MinGroupSequence != 0 && (sr.subscription.MinGroupSequence < update.MinGroupSequence) {
		sr.subscription.MinGroupSequence = update.MinGroupSequence
	}

	// Update the max group sequence
	if update.MaxGroupSequence != 0 && (sr.subscription.MaxGroupSequence > update.MaxGroupSequence) {
		sr.subscription.SubscriberPriority = update.SubscriberPriority
	}

	// Update the parameters
	for k, v := range update.Parameters {
		sr.subscription.Parameters.Add(k, v)
	}

	// Update the delivery timeout
	if update.DeliveryTimeout != 0 {
		sr.subscription.DeliveryTimeout = update.DeliveryTimeout
	}
}

func (sr SubscribeReceiver) ReceiveUpdate() (SubscribeUpdate, error) {
	return readSubscribeUpdate(sr.stream)
}

func (sr SubscribeReceiver) Inform(i Info) {
	slog.Debug("Accepting the subscription")

	im := message.InfoMessage{
		PublisherPriority:   message.PublisherPriority(i.PublisherPriority),
		LatestGroupSequence: message.GroupSequence(i.LatestGroupSequence),
		GroupOrder:          message.GroupOrder(i.GroupOrder),
		GroupExpires:        i.GroupExpires,
	}

	err := im.Encode(sr.stream)
	if err != nil {
		slog.Error("failed to accept the Subscription", slog.String("error", err.Error()))
		sr.CancelRead(err)
		return
	}

	slog.Info("Accepted the subscription")
}

// TODO: rename this to CancelReceive
func (sr SubscribeReceiver) CancelRead(err error) {
	slog.Debug("canceling a subscription", slog.Any("subscription", sr.subscription))

	if err == nil {
		sr.Close()
		return
	}

	var code moq.StreamErrorCode

	var strerr moq.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		suberr, ok := err.(SubscribeError)
		if ok {
			code = moq.StreamErrorCode(suberr.SubscribeErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	sr.stream.CancelRead(code)
	sr.stream.CancelWrite(code)

	slog.Debug("Rejected a subscription", slog.String("error", err.Error()))
}

func (sr SubscribeReceiver) Close() {
	slog.Info("Closing a Subscrbe Receiver", slog.Any("subscription", sr.subscription))
	err := sr.stream.Close()
	if err != nil {
		slog.Debug("catch an error when closing a Subscribe Stream", slog.String("error", err.Error()))
	}
}

func readSubscription(r moq.Stream) (Subscription, error) {
	var sm message.SubscribeMessage
	err := sm.Decode(r)
	if err != nil {
		slog.Debug("failed to read a SUBSCRIBE message", slog.String("error", err.Error()))
		return Subscription{}, err
	}

	return Subscription{
		subscribeID:        SubscribeID(sm.SubscribeID),
		TrackPath:          sm.TrackPath,
		SubscriberPriority: SubscriberPriority(sm.SubscriberPriority),
		GroupOrder:         GroupOrder(sm.GroupOrder),
		MinGroupSequence:   GroupSequence(sm.MinGroupSequence),
		MaxGroupSequence:   GroupSequence(sm.MaxGroupSequence),
		Parameters:         Parameters(sm.Parameters),
	}, nil
}

type SubscribeUpdate struct {
	SubscriberPriority SubscriberPriority
	GroupOrder         GroupOrder
	GroupExpires       time.Duration
	MinGroupSequence   GroupSequence
	MaxGroupSequence   GroupSequence

	/*
	 * Parameters
	 */
	Parameters Parameters

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
		SubscriberPriority: SubscriberPriority(sum.SubscriberPriority),
		GroupOrder:         GroupOrder(sum.GroupOrder),
		GroupExpires:       sum.GroupExpires,
		MinGroupSequence:   GroupSequence(sum.MinGroupSequence),
		MaxGroupSequence:   GroupSequence(sum.MaxGroupSequence),
		Parameters:         Parameters(sum.Parameters),
		DeliveryTimeout:    timeout,
	}, nil
}
