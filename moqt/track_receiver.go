package moqt

import (
	"context"
	"sync"
)

func newTrackReceiver(substr *sendSubscribeStream) *trackReceiver {
	track := &trackReceiver{
		substr:   substr,
		queuedCh: make(chan struct{}, 1),
		queue:    make([]*receiveGroupStream, 0, 1<<4),
		dequeued: make(map[*receiveGroupStream]struct{}),
	}

	// Close the receiver when the subscribe stream context is done.
	go func() {
		<-substr.ctx.Done()
		_ = track.Close()
	}()

	return track
}

var _ TrackReader = (*trackReceiver)(nil)

type trackReceiver struct {
	queue    []*receiveGroupStream
	queuedCh chan struct{}
	mu       sync.Mutex

	once sync.Once

	dequeued map[*receiveGroupStream]struct{}

	substr *sendSubscribeStream
}

func (r *trackReceiver) AcceptGroup(ctx context.Context) (GroupReader, error) {
	for {
		r.mu.Lock()
		if len(r.queue) > 0 {
			next := r.queue[0]

			r.queue = r.queue[1:]

			if next == nil {
				r.mu.Unlock()
				continue
			}

			r.dequeued[next] = struct{}{}
			go func() {
				<-next.doneCh
				r.mu.Lock()
				defer r.mu.Unlock()

				delete(r.dequeued, next)
			}()

			r.mu.Unlock()
			return next, nil
		}

		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-r.substr.ctx.Done():
			return nil, r.substr.ctx.Err()
		case <-r.queuedCh:
		}
	}
}

func (r *trackReceiver) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.once.Do(func() {
		close(r.queuedCh)
	})

	return r.substr.close()
}

func (r *trackReceiver) CloseWithError(code SubscribeErrorCode) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.once.Do(func() {
		close(r.queuedCh)
	})

	for _, stream := range r.queue {
		stream.CancelRead(SubscribeCanceledErrorCode)
	}
	r.queue = nil

	for stream := range r.dequeued {
		stream.CancelRead(SubscribeCanceledErrorCode)
	}
	r.dequeued = nil

	return r.substr.closeWithError(code)
}

func (r *trackReceiver) enqueueGroup(stream *receiveGroupStream) {
	if stream == nil {
		return
	}

	if !r.substr.SubscribeConfig().IsInRange(stream.GroupSequence()) {
		stream.CancelRead(OutOfRangeErrorCode)
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.queue = append(r.queue, stream)

	// Send a notification (non-blocking)
	select {
	case r.queuedCh <- struct{}{}:
	default:
	}
}
