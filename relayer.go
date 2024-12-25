package moqt

import (
	"sync"
)

func Relay(RelayManager *RelayManager, sess ServerSession) {
	//TODO
}

type relayer interface {
	Interest(Interest) (*SentInterest, error)

	Subscribe(Subscription) (*SentSubscription, error)
	Unsubscribe(*SentSubscription)

	Fetch(Fetch) (DataReceiveStream, error)

	RequestInfo(InfoRequest) (Info, error)
}

var _ relayer = (*Relayer)(nil)

func NewRelayer(upstream Subscriber, buffSize int) *Relayer {
	return &Relayer{
		upstream:    upstream,
		downstreams: make([]*Publisher, 0),
		BufferSize:  buffSize,
	}
}

type Relayer struct {
	//trackPath string

	upstream Subscriber

	downstreams []*Publisher
	dsMu        sync.RWMutex

	BufferSize int

	//CacheManager CacheManager
}

func (r *Relayer) Interest(interest Interest) (*SentInterest, error) {
	return r.upstream.Interest(interest)
}

func (r *Relayer) Subscribe(sub Subscription) (*SentSubscription, error) {
	return r.upstream.Subscribe(sub)
}

func (r *Relayer) Unsubscribe(sub *SentSubscription) {
	r.upstream.Unsubscribe(sub)
}

func (r *Relayer) Fetch(fetch Fetch) (DataReceiveStream, error) {
	return r.upstream.Fetch(fetch)
}

func (r *Relayer) RequestInfo(req InfoRequest) (Info, error) {
	return r.upstream.RequestInfo(req)
}

func (r *Relayer) addDownstream(p *Publisher) {
	r.dsMu.Lock()
	defer r.dsMu.Unlock()

	r.downstreams = append(r.downstreams, p)
}

func (r *Relayer) removeDownstream(p *Publisher) {
	r.dsMu.Lock()
	defer r.dsMu.Unlock()

	for i, downstream := range r.downstreams {
		if downstream == p {
			r.downstreams = append(r.downstreams[:i], r.downstreams[i+1:]...)
			break
		}
	}
}
