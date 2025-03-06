package moqt

import (
	"fmt"
)

type SubscribeUpdate struct {
	TrackPriority    TrackPriority
	GroupOrder       GroupOrder
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence

	/*
	 * SubscribeParameters
	 */
	// SubscribeParameters Parameters
}

func (su SubscribeUpdate) String() string {
	return fmt.Sprintf("SubscribeUpdate: { TrackPriority: %d, GroupOrder: %d, MinGroupSequence: %d, MaxGroupSequence: %d }",
		su.TrackPriority, su.GroupOrder, su.MinGroupSequence, su.MaxGroupSequence)
}
