package moqt

import (
	"sync"
)

type relayer interface {
}

var _ relayer = (*Relayer)(nil)

// func newRelayer(path string, upstream ServerSession) *Relayer {
// 	return &Relayer{
// 		TrackPath:   path,
// 		upstream:    upstream,
// 		downstreams: make([]ServerSession, 0),
// 		// BufferSize: 1,
// 	}
// }

type Relayer struct {
	TrackPath string

	upstream *Subscriber

	downstreams []*Publisher
	dsMu        sync.RWMutex

	BufferSize int

	//CacheManager CacheManager
}
