package moqt

import (
	"context"
	"errors"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/quic"
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
	trackMu  sync.Mutex

	dequeued map[*GroupReader]struct{}

	onCloseTrackFunc func()
}

func (r *TrackReader) AcceptGroup(ctx context.Context) (*GroupReader, error) {
	for {
		r.trackMu.Lock()
		if len(r.queueing) > 0 {
			next := r.queueing[0]

			r.queueing = r.queueing[1:]

			r.trackMu.Unlock()

			var group *GroupReader
			group = newGroupReader(next.sequence, next.stream,
				func() { r.removeGroup(group) })

			r.addGroup(group)

			return group, nil
		}

		trackCtx := r.Context()

		if trackCtx.Err() != nil {
			r.trackMu.Unlock()
			reason := context.Cause(trackCtx)
			var strErr *quic.StreamError
			if errors.As(reason, &strErr) {
				return nil, &SubscribeError{StreamError: strErr}
			}

			var appErr *quic.ApplicationError
			if errors.As(reason, &appErr) {
				return nil, &SessionError{ApplicationError: appErr}
			}
			return nil, reason
		}

		queueCh := r.queuedCh
		r.trackMu.Unlock()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-trackCtx.Done():
			reason := context.Cause(trackCtx)
			var strErr *quic.StreamError
			if errors.As(reason, &strErr) {
				return nil, &SubscribeError{
					StreamError: strErr,
				}
			}

			var appErr *quic.ApplicationError
			if errors.As(reason, &appErr) {
				return nil, &SessionError{
					ApplicationError: appErr,
				}
			}
			return nil, reason
		case <-queueCh:
		}
	}
}

func (r *TrackReader) Close() error {
	r.trackMu.Lock()
	defer r.trackMu.Unlock()

	// Cancel all pending groups first
	errCode := quic.StreamErrorCode(SubscribeCanceledErrorCode)
	for _, entry := range r.queueing {
		entry.stream.CancelRead(errCode)
	}
	r.queueing = nil

	// Cancel all dequeued groups
	for stream := range r.dequeued {
		stream.CancelRead(SubscribeCanceledErrorCode)
	}
	r.dequeued = nil

	if r.queuedCh != nil {
		close(r.queuedCh)
		r.queuedCh = nil
	}

	r.onCloseTrackFunc()

	return r.sendSubscribeStream.close()
}

func (r *TrackReader) CloseWithError(code SubscribeErrorCode) error {
	r.trackMu.Lock()
	defer r.trackMu.Unlock()

	// Cancel all pending groups first
	errCode := quic.StreamErrorCode(code)
	for _, entry := range r.queueing {
		entry.stream.CancelRead(errCode)
	}
	r.queueing = nil

	// Cancel all dequeued groups
	for stream := range r.dequeued {
		stream.CancelRead(SubscribeCanceledErrorCode)
	}
	r.dequeued = nil

	if r.queuedCh != nil {
		close(r.queuedCh)
		r.queuedCh = nil
	}

	r.onCloseTrackFunc()

	return r.sendSubscribeStream.closeWithError(code)
}

func (r *TrackReader) Update(config *TrackConfig) error {
	if config == nil {
		return errors.New("subscribe config cannot be nil")
	}

	return r.sendSubscribeStream.UpdateSubscribe(config)
}

func (r *TrackReader) TrackConfig() *TrackConfig {
	return r.sendSubscribeStream.TrackConfig()
}

func (r *TrackReader) enqueueGroup(sequence GroupSequence, stream quic.ReceiveStream) {
	if stream == nil {
		return
	}

	r.trackMu.Lock()
	defer r.trackMu.Unlock()

	if r.Context().Err() != nil || r.queueing == nil {
		stream.CancelRead(quic.StreamErrorCode(SubscribeCanceledErrorCode))
		return
	}

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
	r.trackMu.Lock()
	defer r.trackMu.Unlock()

	r.dequeued[group] = struct{}{}
}

func (r *TrackReader) removeGroup(group *GroupReader) {
	r.trackMu.Lock()
	defer r.trackMu.Unlock()

	delete(r.dequeued, group)
}
