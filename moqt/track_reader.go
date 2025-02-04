package moqt

type TrackReader interface {
	GetGroup(GroupSequence) GroupReader
	DequeueGroup() GroupReader
	Closed() bool
}
