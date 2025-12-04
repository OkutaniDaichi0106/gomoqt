package moqt

import (
	"fmt"
)

// TrackConfig holds subscription parameters for a track. It is used to
// specify the range of group sequences to receive and the delivery priority
// for the track.
type TrackConfig struct {
	TrackPriority TrackPriority
}

func (sc TrackConfig) String() string {
	return fmt.Sprintf("{ track_priority: %d }",
		sc.TrackPriority)
}
