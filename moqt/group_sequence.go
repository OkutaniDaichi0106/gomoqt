package moqt

import "fmt"

const (
	// MinGroupSequence is the smallest valid sequence number for a group.
	MinGroupSequence GroupSequence = 0
	// MaxGroupSequence is the largest valid sequence number for a group.
	// It is set to 2^62 - 1, the largest sequence number that can be used in a group.
	MaxGroupSequence GroupSequence = 0x3FFFFFFFFFFFFFFF
)

type GroupSequence uint64

// String returns the string representation of the group sequence number.
func (gs GroupSequence) String() string {
	return fmt.Sprintf("%d", gs)
}

// Next returns the next sequence number in the group sequence.
// If the current sequence is at the maximum, it wraps around to the minimum.
func (gs GroupSequence) Next() GroupSequence {
	if gs == MinGroupSequence {
		return 1
	}

	if gs == MaxGroupSequence {
		// Wrap to the first sequence value
		return MinGroupSequence
	}

	return gs + 1
}
