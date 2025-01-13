package moqt

import (
	"log/slog"
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
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence

	SubscribeParameters Parameters
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

	return SubscribeID(sm.SubscribeID), subscription, nil
}

func writeSubscription(w transport.Stream, id SubscribeID, subscription Subscription) error {
	// Set parameters
	if subscription.SubscribeParameters == nil {
		subscription.SubscribeParameters = make(Parameters)
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
