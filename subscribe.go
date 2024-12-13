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

type Subscription struct {
	Track
	subscribeID SubscribeID
	//TrackPath          string
	SubscriberPriority Priority
	//GroupOrder         GroupOrder
	//GroupExpires     time.Duration
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence

	/*
	 * Parameters
	 */
	Parameters Parameters

	DeliveryTimeout time.Duration //TODO

	/*
	 * Not in wire
	 */
}

func (s Subscription) getGroup(seq GroupSequence, priority Priority) Group {
	return Group{
		subscribeID:       s.subscribeID,
		groupSequence:     seq,
		PublisherPriority: priority,
	}
}

func newSubscribeSendStream(subscription Subscription, stream moq.Stream) *subscribeSendStream {
	return &subscribeSendStream{
		subscription: subscription,
		stream:       stream,
	}
}

type subscribeSendStream struct {
	subscription Subscription
	stream       moq.Stream
	mu           sync.Mutex
}

/*
 *
 */

type subscribeReceiveStream struct {
	subscription Subscription
	stream       moq.Stream
	mu           sync.Mutex
}

func (sr *subscribeReceiveStream) ReceiveUpdate() (SubscribeUpdate, error) {
	return readSubscribeUpdate(sr.stream)
}

func (sr *subscribeReceiveStream) Inform(info Info) {
	slog.Debug("Accepting the subscription")

	im := message.InfoMessage{
		PublisherPriority:   message.Priority(info.PublisherPriority),
		LatestGroupSequence: message.GroupSequence(info.LatestGroupSequence),
		GroupOrder:          message.GroupOrder(info.GroupOrder),
		GroupExpires:        info.GroupExpires,
	}

	err := im.Encode(sr.stream)
	if err != nil {
		slog.Error("failed to inform track status", slog.String("error", err.Error()))
		return
	}

	slog.Info("Informed", slog.Any("info", info))
}

func (srs *subscribeReceiveStream) CloseWithError(err error) {
	if err == nil {
		srs.Close()
	}

	// TODO:

	var code moq.StreamErrorCode

	var strerr moq.StreamError
	if errors.As(err, &strerr) {
		code = strerr.StreamErrorCode()
	} else {
		var ok bool
		feterr, ok := err.(FetchError)
		if ok {
			code = moq.StreamErrorCode(feterr.FetchErrorCode())
		} else {
			code = ErrInternalError.StreamErrorCode()
		}
	}

	srs.stream.CancelRead(code)
	srs.stream.CancelWrite(code)

	slog.Info("rejcted the fetch request")
}

func (sr *subscribeReceiveStream) Close() {
	slog.Info("Closing a subscrbe receive stream", slog.Any("subscription", sr.subscription))
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
		subscribeID: SubscribeID(sm.SubscribeID),
		Track: Track{
			TrackPath:    sm.TrackPath,
			GroupOrder:   GroupOrder(sm.GroupOrder),
			GroupExpires: sm.GroupExpires,
		},
		SubscriberPriority: Priority(sm.SubscriberPriority),
		MinGroupSequence:   GroupSequence(sm.MinGroupSequence),
		MaxGroupSequence:   GroupSequence(sm.MaxGroupSequence),
		Parameters:         Parameters(sm.Parameters),
	}, nil
}

type SubscribeUpdate struct {
	SubscriberPriority Priority
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
		SubscriberPriority: Priority(sum.SubscriberPriority),
		GroupOrder:         GroupOrder(sum.GroupOrder),
		GroupExpires:       sum.GroupExpires,
		MinGroupSequence:   GroupSequence(sum.MinGroupSequence),
		MaxGroupSequence:   GroupSequence(sum.MaxGroupSequence),
		Parameters:         Parameters(sum.Parameters),
		DeliveryTimeout:    timeout,
	}, nil
}
