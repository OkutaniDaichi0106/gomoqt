package moqt

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type SubscribeUpdate struct {
	TrackPriority    TrackPriority
	GroupOrder       GroupOrder
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence

	/*
	 * SubscribeParameters
	 */
	SubscribeParameters Parameters
}

func (su SubscribeUpdate) String() string {
	return fmt.Sprintf("SubscribeUpdate: { TrackPriority: %d, GroupOrder: %d, MinGroupSequence: %d, MaxGroupSequence: %d, SubscribeParameters: %s }",
		su.TrackPriority, su.GroupOrder, su.MinGroupSequence, su.MaxGroupSequence, su.SubscribeParameters.String())
}

func readSubscribeUpdate(r io.Reader) (SubscribeUpdate, error) {
	var sum message.SubscribeUpdateMessage
	if _, err := sum.Decode(r); err != nil {
		slog.Error("failed to decode SUBSCRIBE_UPDATE message",
			"error", err)
		return SubscribeUpdate{}, fmt.Errorf("decode SUBSCRIBE_UPDATE: %w", err)
	}

	return SubscribeUpdate{
		TrackPriority:       TrackPriority(sum.TrackPriority),
		GroupOrder:          GroupOrder(sum.GroupOrder),
		MinGroupSequence:    GroupSequence(sum.MinGroupSequence),
		MaxGroupSequence:    GroupSequence(sum.MaxGroupSequence),
		SubscribeParameters: Parameters{paramMap: sum.SubscribeUpdateParameters},
	}, nil
}

func writeSubscribeUpdate(w io.Writer, update SubscribeUpdate) error {
	// Initialize parameters if nil
	if update.SubscribeParameters.paramMap == nil {
		update.SubscribeParameters = NewParameters()
	}

	sum := message.SubscribeUpdateMessage{
		TrackPriority:             message.TrackPriority(update.TrackPriority),
		GroupOrder:                message.GroupOrder(update.GroupOrder),
		MinGroupSequence:          message.GroupSequence(update.MinGroupSequence),
		MaxGroupSequence:          message.GroupSequence(update.MaxGroupSequence),
		SubscribeUpdateParameters: update.SubscribeParameters.paramMap,
	}

	if _, err := sum.Encode(w); err != nil {
		return fmt.Errorf("encode SUBSCRIBE_UPDATE: %w", err)
	}
	return nil
}

func updateSubscription(subscription SubscribeConfig, update SubscribeUpdate) (SubscribeConfig, error) {
	// Validate sequence ranges
	if subscription.MinGroupSequence > update.MinGroupSequence ||
		subscription.MaxGroupSequence < update.MaxGroupSequence {
		return subscription, ErrInvalidRange
	}

	// Update all fields
	subscription.TrackPriority = update.TrackPriority
	subscription.GroupOrder = update.GroupOrder
	subscription.MinGroupSequence = update.MinGroupSequence
	subscription.MaxGroupSequence = update.MaxGroupSequence

	// Update parameters
	if subscription.SubscribeParameters.paramMap == nil {
		subscription.SubscribeParameters = NewParameters()
	}
	for k, v := range update.SubscribeParameters.paramMap {
		subscription.SubscribeParameters.paramMap[k] = v
	}

	return subscription, nil
}
