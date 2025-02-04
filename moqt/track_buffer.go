package moqt

type TrackBuffer interface {
	//
	LatestGroupSequence() GroupSequence

	// Producer
	CreateGroup(GroupSequence) GroupWriter

	// Relayer
	EnqueueGroup(gr GroupReader)
	DequeueGroup() GroupReader

	// Consumer
	GetGroup(GroupSequence) GroupReader

	//
	Close() error
	CloseWithError(error) error
	Closed() bool
}

// var _ trackBuffer = (*TrackBuffer)(nil)

// func NewTrackBuffer(trackPath []string) TrackBuffer {
// 	return TrackBuffer{
// 		TrackPath: trackPath,
// 		groups:    make(map[GroupSequence]*groupBuffer),
// 		cond:      sync.NewCond(&sync.Mutex{}),
// 	}
// }

// type TrackBuffer struct {
// 	TrackPath []string
// 	groups    map[GroupSequence]*groupBuffer
// 	cond      *sync.Cond
// }

// func (tb *TrackBuffer) ReadGroup(gr GroupReader) {
// 	buffer := newGroupBuffer(gr)
// 	tb.groups[gr.GroupSequence()] = buffer
// }

// func (tb *TrackBuffer) WriteGroup(gr GroupWriter) {
// 	buffer := tb.groups[gr.GroupSequence()]
// 	if buffer == nil {
// 		return
// 	}
// 	buffer.Relay(gr)
// }

// type trackBufer interface {
// 	// 1. Add a new group and serve some data
// 	// 2. Dequeue a group
// }

type TrackRelayer interface {
	Enqueue(gr GroupReader)
	Dequeue() GroupReader
}
