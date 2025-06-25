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
	GroupSequenceNotSpecified GroupSequence = 0
	GroupSequenceLatest       GroupSequence = 0
	GroupSequenceLargest      GroupSequence = 0
	GroupSequenceFirst        GroupSequence = 1
	MaxGroupSequence          GroupSequence = 0x3FFFFFFFFFFFFFFF // 2^62 - 1, the largest sequence number that can be used in a group
)

func (gs GroupSequence) String() string {
	return fmt.Sprintf("%d", gs)
}

func (gs GroupSequence) Next() GroupSequence {
	if gs == GroupSequenceNotSpecified {
		return 1
	}

	if gs == MaxGroupSequence {
		// WARN: Is this behavior acceptable?
		return 1
	}

	return gs + 1
}
