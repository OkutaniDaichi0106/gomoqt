package moqt

import (
	"fmt"
)

type SubscribeID uint64

func (id SubscribeID) String() string {
	return fmt.Sprintf("SubscribeID: %d", id)
}

type SubscribeConfig struct {
	/*
	 * Required
	 */
	TrackPath TrackPath

	/*
	 * Optional
	 */
	TrackPriority    TrackPriority
	GroupOrder       GroupOrder
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence
}

func (sc SubscribeConfig) String() string {
	return fmt.Sprintf("SubscribeConfig: { TrackPath: %s, TrackPriority: %d, GroupOrder: %d, MinGroupSequence: %d, MaxGroupSequence: %d }",
		sc.TrackPath.String(), sc.TrackPriority, sc.GroupOrder, sc.MinGroupSequence, sc.MaxGroupSequence)
}
