package moqt

import "sync"

func NewTrackBuffer() TrackBuffer {
	return TrackBuffer{
		groups: make(map[GroupSequence]*GroupRelayer),
		cond:   sync.NewCond(&sync.Mutex{}),
	}
}

type TrackBuffer struct {
	groups map[GroupSequence]*GroupRelayer
	cond   *sync.Cond
}

func (tb *TrackBuffer) AddGroup(seq GroupSequence) {

}
