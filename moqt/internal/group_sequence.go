package internal

import "fmt"

const (
	GroupSequenceNotSpecified GroupSequence = 0
	GroupSequenceLatest       GroupSequence = 0
	GroupSequenceLargest      GroupSequence = 0
	GroupSequenceFirst        GroupSequence = 1
	MaxGroupSequence          GroupSequence = 0x3FFFFFFFFFFFFFFF // 2^62 - 1, the largest sequence number that can be used in a group
)

type GroupSequence uint64

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
