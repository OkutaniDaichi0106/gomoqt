package moqt

import "context"

type TrackReader interface {
	// Get the track path
	TrackPath() TrackPath

	// Get the track priority
	TrackPriority() TrackPriority

	// Get the group order
	GroupOrder() GroupOrder

	// Get the latest group sequence
	LatestGroupSequence() GroupSequence

	// Get the track info
	Info() Info

	// Accept a group
	AcceptGroup(context.Context) (GroupReader, error)

	Close() error

	CloseWithError(error) error
}
