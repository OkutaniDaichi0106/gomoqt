package moqt

import (
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type SendSubscribeStream interface {
	SubscribeID() SubscribeID
	Subscription() Subscription
	UpdateSubscribe(SubscribeUpdate) error
	Unsubscribe() error
	Close() error
	CloseWithError(error) error
}

var _ SendSubscribeStream = (*sendSubscribeStream)(nil)

type sendSubscribeStream struct {
	subscribeID  SubscribeID
	subscription Subscription
	stream       transport.Stream
	mu           sync.Mutex
}

func (ss *sendSubscribeStream) SubscribeID() SubscribeID {
	return ss.subscribeID
}

func (ss *sendSubscribeStream) Subscription() Subscription {
	return ss.subscription
}

func (sss *sendSubscribeStream) UpdateSubscribe(update SubscribeUpdate) error {
	/*
	 * Verify the update
	 */
	// Verify if the new group range is valid
	if update.MinGroupSequence > update.MaxGroupSequence {
		slog.Debug("MinGroupSequence is larger than MaxGroupSequence")
		return ErrInvalidRange
	}
	// Verify if the minimum group sequence become larger
	if sss.subscription.MinGroupSequence > update.MinGroupSequence {
		slog.Debug("the new MinGroupSequence is smaller than the old MinGroupSequence")
		return ErrInvalidRange
	}
	// Verify if the maximum group sequence become smaller
	if sss.subscription.MaxGroupSequence < update.MaxGroupSequence {
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
		SubscribeID:      message.SubscribeID(sss.SubscribeID()),
		TrackPriority:    message.TrackPriority(update.TrackPriority),
		GroupOrder:       message.GroupOrder(update.GroupOrder),
		GroupExpires:     update.GroupExpires,
		MinGroupSequence: message.GroupSequence(update.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(update.MaxGroupSequence),
		Parameters:       message.Parameters(update.SubscribeParameters),
	}
	err := sum.Encode(sss.stream)
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
		return err
	}

	// Receive an INFO message
	info, err := readInfo(sss.stream)
	if err != nil {
		slog.Debug("failed to get an Info")
		return err
	}

	// Update the TrackPriority
	if info.TrackPriority == update.TrackPriority {
		sss.subscription.TrackPriority = info.TrackPriority
	} else {
		slog.Debug("TrackPriority is not updated")
		return ErrPriorityMismatch
	}

	// Update the GroupOrder
	if update.GroupOrder == 0 {
		sss.subscription.GroupOrder = info.GroupOrder
	} else {
		if info.GroupOrder != update.GroupOrder {
			slog.Debug("GroupOrder is not updated")
			return ErrGroupOrderMismatch
		}

		sss.subscription.GroupOrder = update.GroupOrder
	}

	// Update the GroupExpires
	if info.GroupExpires < update.GroupExpires {
		sss.subscription.GroupExpires = info.GroupExpires
	} else {
		sss.subscription.GroupExpires = update.GroupExpires
	}

	// Update the MinGroupSequence and MaxGroupSequence
	sss.subscription.MinGroupSequence = update.MinGroupSequence
	sss.subscription.MaxGroupSequence = update.MaxGroupSequence

	// Update the SubscribeParameters
	sss.subscription.SubscribeParameters = update.SubscribeParameters

	// Update the DeliveryTimeout
	if update.DeliveryTimeout != 0 {
		sss.subscription.DeliveryTimeout = update.DeliveryTimeout
	}

	return nil
}

func (sss *sendSubscribeStream) Unsubscribe() error {
	// TODO: Implement

	return sss.Close()
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

	sss.stream.CancelRead(code)
	sss.stream.CancelWrite(code)

	slog.Debug("closed a subscrbe receive stream", slog.Any("subscription", sss.Subscription()))

	return nil
}

type ReceiveSubscribeStream interface {
	SubscribeID() SubscribeID
	Subscription() Subscription
	CountDataGap(uint64) error
	CloseWithError(error) error
	Close() error
}

var _ ReceiveSubscribeStream = (*receiveSubscribeStream)(nil)

func newReceivedSubscription(stream transport.Stream) (*receiveSubscribeStream, error) {
	id, subscription, err := readSubscription(stream)
	if err != nil {
		slog.Error("failed to get a subscription", slog.String("error", err.Error()))
		return nil, err
	}

	rs := &receiveSubscribeStream{
		subscribeID:  id,
		subscription: subscription,
		stream:       stream,
	}

	// go rs.listenUpdate()

	return rs, nil
}

type receiveSubscribeStream struct {
	subscribeID  SubscribeID
	subscription Subscription
	stream       transport.Stream
	mu           sync.Mutex
}

func (rss *receiveSubscribeStream) SubscribeID() SubscribeID {
	return rss.subscribeID
}

func (rss *receiveSubscribeStream) Subscription() Subscription {
	return rss.subscription
}

func (rss *receiveSubscribeStream) updateLastestGroupSequence(sequence GroupSequence) {
	atomic.StoreUint64((*uint64)(&rss.subscription.latestGroupSequence), uint64(sequence))
}

func (rs *receiveSubscribeStream) CountDataGap(code uint64) error {
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
