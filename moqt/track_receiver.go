package moqt

import (
	"context"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newTrackReceiver(ctx context.Context) *trackReceiver {
	track := &trackReceiver{
		queuedCh: make(chan struct{}, 1),
		queue:    make([]*receiveGroupStream, 0, 1<<4),
		dequeued: make(map[*receiveGroupStream]struct{}),
	}

	go func() {
		<-ctx.Done()
		track.mu.Lock()
		defer track.mu.Unlock()

		// Clear the queue and notify all waiting goroutines
		for _, stream := range track.queue {
			stream.CancelRead(SubscribeCanceledErrorCode)
		}
		for stream := range track.dequeued {
			stream.CancelRead(SubscribeCanceledErrorCode)
		}
		track.queue = nil
		track.dequeued = nil
	}()

	return track
}

var _ TrackReader = (*trackReceiver)(nil)

type trackReceiver struct {
	ctx context.Context

	queue    []*receiveGroupStream
	queuedCh chan struct{}
	mu       sync.Mutex

	dequeued map[*receiveGroupStream]struct{}
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
				<-next.ctx.Done()
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
		case <-r.ctx.Done():
			return nil, r.ctx.Err()
		case <-r.queuedCh:
		}
	}
}

func (r *trackReceiver) enqueueGroup(GroupSequence GroupSequence, stream quic.ReceiveStream) {
	if stream == nil {
		return
	}

	group := newReceiveGroupStream(r.ctx, GroupSequence, stream)

	// if !r.substr.SubscribeConfig().IsInRange(stream.GroupSequence()) {
	// 	stream.CancelRead(OutOfRangeErrorCode)
	// 	return
	// }

	r.mu.Lock()
	defer r.mu.Unlock()

	r.queue = append(r.queue, group)

	// Send a notification (non-blocking)
	select {
	case r.queuedCh <- struct{}{}:
	default:
	}
}
