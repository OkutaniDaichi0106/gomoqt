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
	TrackPriority TrackPriority
	GroupOrder    GroupOrder

	// Parameters
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence

	SubscribeParameters Parameters
}

func (sc SubscribeConfig) String() string {
	return fmt.Sprintf("SubscribeConfig: { TrackPath: %s, TrackPriority: %d, GroupOrder: %d, MinGroupSequence: %d, MaxGroupSequence: %d, SubscribeParameters: %s }",
		sc.TrackPath.String(), sc.TrackPriority, sc.GroupOrder, sc.MinGroupSequence, sc.MaxGroupSequence, sc.SubscribeParameters.String())
}
