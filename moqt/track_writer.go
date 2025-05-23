package moqt

type TrackWriter interface {
	// Create a new group writer
	OpenGroup(GroupSequence) (GroupWriter, error)

	SendGap(Gap) error

	Close() error

	CloseWithError(error) error
}
