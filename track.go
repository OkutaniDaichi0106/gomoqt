package moqt

import (
	"time"
)

type Track struct {
	TrackPath string

	TrackPriority TrackPriority
	GroupOrder    GroupOrder
	GroupExpires  time.Duration

	/*
	 * Parameters
	 */
	AuthorizationInfo string

	DeliveryTimeout time.Duration //TODO

	AnnounceParameters Parameters

	/*
	 *
	 */
	latestGroupSequence GroupSequence
}

func (t *Track) Info() Info {
	return Info{
		TrackPriority:       t.TrackPriority,
		LatestGroupSequence: t.latestGroupSequence,
		GroupOrder:          t.GroupOrder,
		GroupExpires:        t.GroupExpires,
	}
}
