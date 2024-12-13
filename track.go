package moqt

import "time"

func NewTrack()

type Track struct {
	TrackPath string

	/*
	 *
	 */
	//TrackPriority    Priority
	GroupOrder   GroupOrder
	GroupExpires time.Duration
	// MinGroupSequence GroupSequence
	// MaxGroupSequence GroupSequence

	/*
	 * Parameters
	 */
	announceParameters Parameters
	AuthorizationInfo  string

	DeliveryTimeout time.Duration //TODO

	/*
	 *
	 */
	groups map[GroupSequence]Group

	// Subscription
}

func (t Track) Announcement() Announcement {
	return Announcement{
		TrackPath:         t.TrackPath,
		AuthorizationInfo: t.AuthorizationInfo,
		Parameters:        t.announceParameters,
	}
}

func (t Track) Info() Info {
	return Info{
		GroupPriority:       0, // TODO:
		LatestGroupSequence: 0, // TODO:
		GroupOrder:          t.GroupOrder,
		GroupExpires:        t.GroupExpires,
	}
}
