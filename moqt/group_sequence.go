package moqt

import (
	"fmt"
)

/*
 * Sequence number of a group in a track
 * When this is integer more than 1, the number means the sequence number.
 * When this is 0, it indicates the sequence number is currently unknown .
 * 0 is used to specify "the latest sequence number" or "the final sequence number of an open-ended track", "the first sequence number of the default order".
 */
type GroupSequence uint64

const (
	NotSpecifiedGroupSequence GroupSequence = 0
	LatestGroupSequence       GroupSequence = 0
	LargestGroupSequence      GroupSequence = 0
	FirstGroupSequence        GroupSequence = 1
	MaxGroupSequence          GroupSequence = 0xFFFFFFFF
)

func (gs GroupSequence) String() string {
	return fmt.Sprintf("GroupSequence: %d", gs)
}

func (gs GroupSequence) Next() GroupSequence {
	if gs == NotSpecifiedGroupSequence {
		return 1
	}

	if gs == MaxGroupSequence {
		return 1
	}

	return gs + 1
}
