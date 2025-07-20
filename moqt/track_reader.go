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
		queueing: make([]struct {
			sequence GroupSequence
			stream   quic.ReceiveStream
		}, 0, 1<<3),
		dequeued:         make(map[*GroupReader]struct{}),
		onCloseTrackFunc: onCloseTrackFunc,
	}

	return track
}

type TrackReader struct {
	BroadcastPath BroadcastPath
	TrackName     TrackName

	*sendSubscribeStream

	queueing []struct {
		sequence GroupSequence
		stream   quic.ReceiveStream
	}
	queuedCh chan struct{}
	mu       sync.Mutex

	dequeued map[*GroupReader]struct{}

	onCloseTrackFunc func()
}

func (r *TrackReader) AcceptGroup(ctx context.Context) (*GroupReader, error) {
	for {
		r.mu.Lock()
		if len(r.queueing) > 0 {
			next := r.queueing[0]

			r.queueing = r.queueing[1:]

			var group *GroupReader
			group = newReceiveGroupStream(r.ctx, next.sequence, next.stream,
				func() { r.removeGroup(group) })

			r.addGroup(group)

			r.mu.Unlock()
			return group, nil
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
	for _, entry := range r.queueing {
		entry.stream.CancelRead(quic.StreamErrorCode(SubscribeCanceledErrorCode))
	}
	for stream := range r.dequeued {
		stream.CancelRead(SubscribeCanceledErrorCode)
	}

	// Then close the subscribe stream
	err := r.sendSubscribeStream.close()

	r.onCloseTrackFunc()

	r.queueing = nil
	r.dequeued = nil
	r.queuedCh = nil

	return err
}

func (r *TrackReader) CloseWithError(code SubscribeErrorCode) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, entry := range r.queueing {
		entry.stream.CancelRead(quic.StreamErrorCode(SubscribeCanceledErrorCode))
	}
	for stream := range r.dequeued {
		stream.CancelRead(SubscribeCanceledErrorCode)
	}

	r.onCloseTrackFunc()

	r.queueing = nil
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

func (r *TrackReader) enqueueGroup(sequence GroupSequence, stream quic.ReceiveStream) {
	if stream == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	entry := struct {
		sequence GroupSequence
		stream   quic.ReceiveStream
	}{
		sequence: sequence,
		stream:   stream,
	}
	r.queueing = append(r.queueing, entry)

	select {
	case r.queuedCh <- struct{}{}:
	default:
	}
}

func (r *TrackReader) addGroup(group *GroupReader) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.dequeued[group] = struct{}{}
}

func (r *TrackReader) removeGroup(group *GroupReader) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.dequeued, group)
}
