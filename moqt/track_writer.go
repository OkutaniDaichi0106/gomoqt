package moqt

type TrackWriter interface {
	// Get the track path
	TrackPath() TrackPath

	// Get the latest group sequence
	LatestGroupSequence() GroupSequence

	// Get the track info
	Info() Info

	// Create a new group writer
	OpenGroup(GroupSequence) (GroupWriter, error)

	Close() error

	CloseWithError(error) error
}
