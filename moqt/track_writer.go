package moqt

type TrackWriter interface {
	// Create a new group writer
	OpenGroup(GroupSequence) (GroupWriter, error)

	Close() error

	CloseWithError(SubscribeErrorCode) error
}
