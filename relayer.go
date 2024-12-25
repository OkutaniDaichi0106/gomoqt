package moqt

import (
	"sync"
)

func Relay(RelayManager *RelayManager, sess ServerSession) {
	//TODO
}

type relayer interface {
	AddDownstream(*Publisher)
}

var _ relayer = (*Relayer)(nil)

func NewRelayer(trackPath string, upstream Subscriber, buffSize int) *Relayer {
	return &Relayer{
		trackPath:   trackPath,
		upstream:    upstream,
		downstreams: make([]*Publisher, 0),
		BufferSize:  buffSize,
	}
}

type Relayer struct {
	trackPath string

	upstream Subscriber

	downstreams []*Publisher
	dsMu        sync.RWMutex

	BufferSize int

	//CacheManager CacheManager
}

func (r *Relayer) AddDownstream(p *Publisher) {
	r.dsMu.Lock()
	defer r.dsMu.Unlock()

	r.downstreams = append(r.downstreams, p)
}

func (r *Relayer) RemoveDownstream(p *Publisher) {
	r.dsMu.Lock()
	defer r.dsMu.Unlock()

	for i, downstream := range r.downstreams {
		if downstream == p {
			r.downstreams = append(r.downstreams[:i], r.downstreams[i+1:]...)
			break
		}
	}
}
