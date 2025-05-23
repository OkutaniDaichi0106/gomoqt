package moqt

import (
	"fmt"
)

type SubscribeConfig struct {
	TrackPriority    TrackPriority
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

func (sc SubscribeConfig) String() string {
	return fmt.Sprintf("SubscribeConfig: { TrackPriority: %d, MinGroupSequence: %d, MaxGroupSequence: %d }",
		sc.TrackPriority, sc.MinGroupSequence, sc.MaxGroupSequence)
}
