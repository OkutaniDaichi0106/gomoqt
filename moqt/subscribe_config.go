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
	if config.MinGroupSequence == GroupSequenceNotSpecified && config.MaxGroupSequence == GroupSequenceNotSpecified {
		return true
	}

	if config.MinGroupSequence == GroupSequenceNotSpecified {
		return seq <= config.MaxGroupSequence
	}

	if config.MaxGroupSequence == GroupSequenceNotSpecified {
		return config.MinGroupSequence <= seq
	}

	return config.MinGroupSequence <= seq && seq <= config.MaxGroupSequence
}

func (sc SubscribeConfig) String() string {
	return fmt.Sprintf("SubscribeConfig: { TrackPriority: %d, MinGroupSequence: %d, MaxGroupSequence: %d }",
		sc.TrackPriority, sc.MinGroupSequence, sc.MaxGroupSequence)
}
