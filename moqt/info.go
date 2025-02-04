package moqt

import (
	"fmt"
)

type Info struct {
	TrackPriority       TrackPriority
	LatestGroupSequence GroupSequence
	GroupOrder          GroupOrder
}

func (i Info) String() string {
	return fmt.Sprintf("Info: { TrackPriority: %d, LatestGroupSequence: %d, GroupOrder: %d }", i.TrackPriority, i.LatestGroupSequence, i.GroupOrder)
}
