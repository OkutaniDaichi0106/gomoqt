package internal

import "github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"

func updateSubscription(sm *message.SubscribeMessage, sum *message.SubscribeUpdateMessage) {
	if sum == nil {
		return
	}

	if sm == nil {
		sm = &message.SubscribeMessage{}
	}

	// Update all fields
	sm.TrackPriority = sum.TrackPriority
	sm.GroupOrder = sum.GroupOrder
	sm.MinGroupSequence = sum.MinGroupSequence
	sm.MaxGroupSequence = sum.MaxGroupSequence

	// Update parameters
	if sum.SubscribeUpdateParameters != nil && sm.SubscribeParameters == nil {
		sm.SubscribeParameters = message.Parameters{}
	}

	for k, v := range sum.SubscribeUpdateParameters {
		sm.SubscribeParameters[k] = v
	}
}
