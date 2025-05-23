package moqt

import (
	"fmt"
)

type SubscribeConfig struct {
	TrackPriority    TrackPriority
	GroupOrder       GroupOrder
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence
}

func (config *SubscribeConfig) IsInRange(seq GroupSequence) bool {
	if config.MinGroupSequence == NotSpecifiedGroupSequence && config.MaxGroupSequence == NotSpecifiedGroupSequence {
		return true
	}

	if config.MinGroupSequence == NotSpecifiedGroupSequence {
		return seq <= config.MaxGroupSequence
	}

	if config.MaxGroupSequence == NotSpecifiedGroupSequence {
		return config.MinGroupSequence <= seq
	}

	return config.MinGroupSequence <= seq && seq <= config.MaxGroupSequence
}

func (config *SubscribeConfig) Update(new *SubscribeConfig) {
	//
	config.TrackPriority = new.TrackPriority

	//
	if new.GroupOrder != GroupOrderDefault {
		config.GroupOrder = new.GroupOrder
	}

	if new.MinGroupSequence != 0 {
		config.MinGroupSequence = new.MinGroupSequence
	}
	if new.MaxGroupSequence != 0 {
		config.MaxGroupSequence = new.MaxGroupSequence
	}
}

func (sc SubscribeConfig) String() string {
	return fmt.Sprintf("SubscribeConfig: { TrackPriority: %d, GroupOrder: %d, MinGroupSequence: %d, MaxGroupSequence: %d }",
		sc.TrackPriority, sc.GroupOrder, sc.MinGroupSequence, sc.MaxGroupSequence)
}
