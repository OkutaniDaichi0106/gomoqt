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
	NotSpecified  GroupSequence = 0
	FirstSequence GroupSequence = 1
	MaxSequence   GroupSequence = 0xFFFFFFFF
)

func (gs GroupSequence) String() string {
	return fmt.Sprintf("GroupSequence: %d", gs)
}

func (gs GroupSequence) IsInRange(min, max GroupSequence) bool {
	if min == NotSpecified && max == NotSpecified {
		return true
	}

	if min == NotSpecified {
		return gs <= max
	}

	if max == NotSpecified {
		return min <= gs
	}

	return min <= gs && gs <= max
}

func (gs GroupSequence) Next() GroupSequence {
	if gs == NotSpecified {
		return 1
	}

	if gs == MaxSequence {
		return 1
	}

	return gs + 1
}
