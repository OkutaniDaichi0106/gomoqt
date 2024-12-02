package moqt

import (
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type SubscribeID uint64

type GroupOrder byte

const (
	DEFAULT    GroupOrder = 0x0
	ASCENDING  GroupOrder = 0x1
	DESCENDING GroupOrder = 0x2
)

type Subscription struct {
	subscribeID        SubscribeID
	TrackNamespace     string
	TrackName          string
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

type SubscribeWriter struct {
	stream       moq.Stream
	subscription Subscription
	mu           sync.RWMutex
}

func (w *SubscribeWriter) Update(update SubscribeUpdate) (Info, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	old := w.subscription
	slog.Debug("trying to update", slog.Any("subscription", w.subscription), slog.Any("to", update))

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

func (s *SubscribeWriter) Unsubscribe(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err == nil {
		err := s.stream.Close()
		if err != nil {
			slog.Error("failed to close a Subscribe Stream", slog.String("error", err.Error()))
		}
		return
	}

	suberr, ok := err.(SubscribeError)
	if !ok {
		suberr = ErrInternalError
	}

	s.stream.CancelWrite(moq.StreamErrorCode(suberr.SubscribeErrorCode()))
	s.stream.CancelRead(moq.StreamErrorCode(suberr.SubscribeErrorCode()))
}

func (s *SubscribeWriter) Subscription() Subscription {
	return s.subscription
}

/*
 *
 */

type SubscribeHandler interface {
	HandleSubscribe(Subscription, *Info, SubscribeResponceWriter)
}

type SubscribeResponceWriter struct {
	stream moq.Stream
}

func (w SubscribeResponceWriter) Accept(i Info) {
	slog.Debug("Accepting the subscription")

	im := message.InfoMessage{
		PublisherPriority:   message.PublisherPriority(i.PublisherPriority),
		LatestGroupSequence: message.GroupSequence(i.LatestGroupSequence),
		GroupOrder:          message.GroupOrder(i.GroupOrder),
		GroupExpires:        i.GroupExpires,
	}

	err := im.Encode(w.stream)
	if err != nil {
		slog.Error("failed to accept the Subscription", slog.String("error", err.Error()))
	}

	slog.Info("Accepted the subscription")
}

func (w SubscribeResponceWriter) Reject(err error) {
	slog.Debug("Rejecting the Subscription")

	if err == nil {
		err := w.stream.Close()
		if err != nil {
			slog.Debug("failed to close a Subscribe Stream gracefully", slog.String("error", err.Error()))
		}

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

	w.stream.CancelRead(code)
	w.stream.CancelWrite(code)

	slog.Debug("Rejected a subscription", slog.String("error", err.Error()))
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
		TrackNamespace:     sm.TrackNamespace,
		TrackName:          sm.TrackName,
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

func readSubscribeUpdate(old Subscription, r io.Reader) (Subscription, error) {

	// Read a SUBSCRIBE_UPDATE message
	var sum message.SubscribeUpdateMessage
	err := sum.Decode(r)
	if err != nil {
		slog.Debug("failed to read a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
		return Subscription{}, err
	}

	new := Subscription{
		subscribeID:        old.subscribeID,
		TrackNamespace:     old.TrackNamespace,
		TrackName:          old.TrackName,
		Parameters:         Parameters(sum.Parameters),
		SubscriberPriority: SubscriberPriority(sum.SubscriberPriority),
		GroupOrder:         GroupOrder(sum.GroupOrder),
		GroupExpires:       sum.GroupExpires,
	}

	return new, nil
}
