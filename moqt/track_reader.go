package moqt

import (
	"context"
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

func newTrackReader(broadcastPath BroadcastPath, trackName TrackName, subscribeStream *sendSubscribeStream, onCloseTrackFunc func()) *TrackReader {
	track := &TrackReader{
		BroadcastPath:       broadcastPath,
		TrackName:           trackName,
		sendSubscribeStream: subscribeStream,
		queuedCh:            make(chan struct{}, 1),
		queue:               make([]*receiveGroupStream, 0, 1<<4),
		dequeued:            make(map[*receiveGroupStream]struct{}),
		onCloseTrackFunc:    onCloseTrackFunc,
	}

	return track
}

type TrackReader struct {
	BroadcastPath BroadcastPath
	TrackName     TrackName

	*sendSubscribeStream

	queue    []*receiveGroupStream
	queuedCh chan struct{}
	mu       sync.Mutex

	dequeued map[*receiveGroupStream]struct{}

	onCloseTrackFunc func()
}

func (r *TrackReader) AcceptGroup(ctx context.Context) (GroupReader, error) {
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

func (r *TrackReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Cancel all active groups first
	for _, stream := range r.queue {
		stream.CancelRead(SubscribeCanceledErrorCode)
	}
	for stream := range r.dequeued {
		stream.CancelRead(SubscribeCanceledErrorCode)
	}

	// Then close the subscribe stream
	err := r.sendSubscribeStream.close()

	r.onCloseTrackFunc()

	r.queue = nil
	r.dequeued = nil
	r.queuedCh = nil

	return err
}

func (r *TrackReader) CloseWithError(code SubscribeErrorCode) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, stream := range r.queue {
		stream.CancelRead(SubscribeCanceledErrorCode)
	}
	for stream := range r.dequeued {
		stream.CancelRead(SubscribeCanceledErrorCode)
	}

	r.onCloseTrackFunc()

	r.queue = nil
	r.dequeued = nil
	r.queuedCh = nil

	r.sendSubscribeStream.closeWithError(code)
}

func (r *TrackReader) Update(config *TrackConfig) error {
	if config == nil {
		return errors.New("subscribe config cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.sendSubscribeStream.UpdateSubscribe(config)
}

func (r *TrackReader) TrackConfig() *TrackConfig {
	return r.sendSubscribeStream.TrackConfig()
}

func (r *TrackReader) enqueueGroup(GroupSequence GroupSequence, stream quic.ReceiveStream) {
	if stream == nil {
		return
	}

	group := newReceiveGroupStream(r.ctx, GroupSequence, stream)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.queue = append(r.queue, group)

	// Send a notification (non-blocking)
	select {
	case r.queuedCh <- struct{}{}:
	default:
	}
}
