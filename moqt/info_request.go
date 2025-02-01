package moqt

import (
	"fmt"
)

type InfoRequest struct {
	TrackPath []string
}

func (ir InfoRequest) String() string {
	return fmt.Sprintf("InfoRequest: { TrackPath: %s }", TrackPartsString(ir.TrackPath))
}
