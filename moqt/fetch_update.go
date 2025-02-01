package moqt

import "fmt"

type FetchUpdate struct {
	TrackPriority TrackPriority
}

func (fu FetchUpdate) String() string {
	return fmt.Sprintf("FetchUpdate: { TrackPriority: %d }", fu.TrackPriority)
}
