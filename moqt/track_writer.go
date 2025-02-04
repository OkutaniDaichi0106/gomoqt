package moqt

type TrackWriter interface {
	CreateGroup(GroupSequence) GroupWriter
	EnqueueGroup(gr GroupReader)
	Close() error
	CloseWithError(error) error
}
