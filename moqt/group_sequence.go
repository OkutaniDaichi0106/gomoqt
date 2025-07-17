package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

/*
 * Sequence number of a group in a track
 * When this is integer more than 1, the number means the sequence number.
 * When this is 0, it indicates the sequence number is currently unknown .
 * 0 is used to specify "the latest sequence number" or "the final sequence number of an open-ended track", "the first sequence number of the default order".
 */
type GroupSequence = protocol.GroupSequence

const (
	GroupSequenceNotSpecified GroupSequence = protocol.GroupSequenceNotSpecified
	GroupSequenceLatest       GroupSequence = protocol.GroupSequenceLatest
	GroupSequenceLargest      GroupSequence = protocol.GroupSequenceLargest
	GroupSequenceFirst        GroupSequence = protocol.GroupSequenceFirst
	MaxGroupSequence          GroupSequence = protocol.MaxGroupSequence
)
