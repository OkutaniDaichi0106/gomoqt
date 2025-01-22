package moqt

import (
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type SubscribeID uint64

// type Subscription struct {
// 	subscribeID SubscribeID
// 	SubscribeConfig
// }

// func (s Subscription) SubscribeID() SubscribeID {
// 	return s.subscribeID
// }

type SubscribeConfig struct {
	/*
	 * Required
	 */
	TrackPath []string

	/*
	 * Optional
	 */
	TrackPriority TrackPriority
	GroupOrder    GroupOrder

	// Parameters
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence

	SubscribeParameters Parameters
}

func readSubscription(r transport.Stream) (SubscribeID, SubscribeConfig, error) {
	var sm message.SubscribeMessage
	err := sm.Decode(r)
	if err != nil {
		slog.Debug("failed to read a SUBSCRIBE message", slog.String("error", err.Error()))
		return 0, SubscribeConfig{}, err
	}

	subscription := SubscribeConfig{
		TrackPath:           sm.TrackPath,
		TrackPriority:       TrackPriority(sm.TrackPriority),
		GroupOrder:          GroupOrder(sm.GroupOrder),
		MinGroupSequence:    GroupSequence(sm.MinGroupSequence),
		MaxGroupSequence:    GroupSequence(sm.MaxGroupSequence),
		SubscribeParameters: Parameters{sm.Parameters},
	}

	return SubscribeID(sm.SubscribeID), subscription, nil
}

func writeSubscription(w transport.Stream, id SubscribeID, subscription SubscribeConfig) error {
	// Set parameters
	if subscription.SubscribeParameters.paramMap == nil {
		subscription.SubscribeParameters = NewParameters()
	}

	// Send a SUBSCRIBE message
	sm := message.SubscribeMessage{
		SubscribeID:      message.SubscribeID(id),
		TrackPath:        subscription.TrackPath,
		TrackPriority:    message.TrackPriority(subscription.TrackPriority),
		GroupOrder:       message.GroupOrder(subscription.GroupOrder),
		MinGroupSequence: message.GroupSequence(subscription.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(subscription.MaxGroupSequence),
		Parameters:       message.Parameters(subscription.SubscribeParameters.paramMap),
	}
	err := sm.Encode(w)
	if err != nil {
		slog.Error("failed to send a SUBSCRIBE message", slog.String("error", err.Error()))
		return err
	}

	return nil
}
