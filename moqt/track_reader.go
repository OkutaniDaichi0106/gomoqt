package moqt

import "context"

type TrackReader interface {
	// Get the SubscribeID
	SubscribeID() SubscribeID

	// Get the track path
	TrackPath() TrackPath

	// Get the track priority
	TrackPriority() TrackPriority

	// Get the group order
	GroupOrder() GroupOrder

	// Get the subscription config
	SubscribeConfig() SubscribeConfig

	// LatestGroupSequence() GroupSequence
	Info() Info

	// Accept a group
	AcceptGroup(context.Context) (GroupReader, error)

	// Update the subscription
	UpdateSubscribe(SubscribeUpdate) error

	//

	// Close the stream
	Close() error

	// Close the stream with an error
	CloseWithError(error) error
}
