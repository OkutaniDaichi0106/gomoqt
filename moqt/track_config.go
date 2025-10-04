package moqt

import (
	"fmt"
)

type TrackConfig struct {
	TrackPriority    TrackPriority
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence
}

func (config *TrackConfig) IsInRange(seq GroupSequence) bool {
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

func (sc TrackConfig) String() string {
	return fmt.Sprintf("{ track_priority: %d, min_group_sequence: %d, max_group_sequence: %d }",
		sc.TrackPriority, sc.MinGroupSequence, sc.MaxGroupSequence)
}
