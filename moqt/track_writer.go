package moqt

type TrackWriter interface {
	WriteInfo(Info)

	// Create a new group writer
	OpenGroup(GroupSequence) (GroupWriter, error)

	Close() error

	CloseWithError(SubscribeErrorCode) error
}
