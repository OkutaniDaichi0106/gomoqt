package moqt

type TrackWriter interface {
	// Get the track path
	TrackPath() TrackPath

	// Get the track priority
	TrackPriority() TrackPriority

	// Get the group order
	GroupOrder() GroupOrder

	// Get the subscription config
	Info() Info

	// Create a new group writer
	OpenGroup(GroupSequence) (GroupWriter, error)

	// Get the subscription config
	SubscribeConfig() SubscribeConfig

	Close() error
	CloseWithError(error) error

	// TODO: Implement
}
