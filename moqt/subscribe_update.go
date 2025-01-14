package moqt

import (
	"io"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

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
}

func readSubscribeUpdate(r io.Reader) (SubscribeUpdate, error) {
	// Read a SUBSCRIBE_UPDATE message
	var sum message.SubscribeUpdateMessage
	err := sum.Decode(r)
	if err != nil {
		slog.Debug("failed to read a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
		return SubscribeUpdate{}, err
	}

	return SubscribeUpdate{
		TrackPriority:       TrackPriority(sum.TrackPriority),
		GroupOrder:          GroupOrder(sum.GroupOrder),
		GroupExpires:        sum.GroupExpires,
		MinGroupSequence:    GroupSequence(sum.MinGroupSequence),
		MaxGroupSequence:    GroupSequence(sum.MaxGroupSequence),
		SubscribeParameters: Parameters{sum.Parameters},
	}, nil
}

func writeSubscribeUpdate(w io.Writer, update SubscribeUpdate) error {
	/*
	 * Send a SUBSCRIBE_UPDATE message
	 */
	// Set parameters
	if update.SubscribeParameters.paramMap == nil {
		update.SubscribeParameters = NewParameters()
	}

	// Send a SUBSCRIBE_UPDATE message
	sum := message.SubscribeUpdateMessage{
		TrackPriority:    message.TrackPriority(update.TrackPriority),
		GroupOrder:       message.GroupOrder(update.GroupOrder),
		GroupExpires:     update.GroupExpires,
		MinGroupSequence: message.GroupSequence(update.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(update.MaxGroupSequence),
		Parameters:       message.Parameters(update.SubscribeParameters.paramMap),
	}
	err := sum.Encode(w)
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func updateSubscription(subscription Subscription, update SubscribeUpdate) (Subscription, error) {
	// Update the Track Priority
	subscription.TrackPriority = update.TrackPriority

	// Update the Group Order
	subscription.GroupOrder = update.GroupOrder

	// Update the Group Expires
	subscription.GroupExpires = update.GroupExpires

	// Update the Min Group Sequence
	if subscription.MinGroupSequence > update.MinGroupSequence {
		return subscription, ErrInvalidRange
	}
	subscription.MinGroupSequence = update.MinGroupSequence

	// Update the Max Group Sequence
	if subscription.MaxGroupSequence < update.MaxGroupSequence {
		return subscription, ErrInvalidRange
	}
	subscription.MaxGroupSequence = update.MaxGroupSequence

	// Update the Subscribe Parameters
	for k, v := range update.SubscribeParameters.paramMap {
		subscription.SubscribeParameters.paramMap[k] = v
	}

	return subscription, nil
}
