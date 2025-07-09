package moqt

import (
	"context"
	"errors"
	"sync"
)

func newTrackSender(ctx context.Context, openGroupFunc func(ctx context.Context, seq GroupSequence) (*sendGroupStream, error)) *trackSender {
	track := &trackSender{
		ctx:           ctx,
		queue:         make(map[*sendGroupStream]struct{}),
		openGroupFunc: openGroupFunc,
	}

	go func() {
		<-ctx.Done()
		track.mu.Lock()
		defer track.mu.Unlock()

		if ctx.Err() != nil {
			for stream := range track.queue {
				stream.CancelWrite(SubscribeCanceledErrorCode)
			}
		} else {
			for stream := range track.queue {
				stream.Close()
			}
		}

		track.queue = nil
	}()
	return track
}

var _ TrackWriter = (*trackSender)(nil)

type trackSender struct {
	ctx context.Context

	acceptFunc func(Info)

	mu    sync.Mutex
	queue map[*sendGroupStream]struct{}

	openGroupFunc func(ctx context.Context, seq GroupSequence) (*sendGroupStream, error)
}

func (s *trackSender) OpenGroup(seq GroupSequence) (GroupWriter, error) {
	if seq == 0 {
		return nil, errors.New("group sequence must not be zero")
	}

	if err := s.ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.acceptFunc(Info{}) // TODO:

	group, err := s.openGroupFunc(s.ctx, seq)
	if err != nil {
		return nil, err
	}

	if s.queue == nil {
		return nil, errors.New("subscription was canceled")
	}

	s.queue[group] = struct{}{}
	go func() {
		<-group.ctx.Done()
		s.mu.Lock()
		delete(s.queue, group)
		s.mu.Unlock()
	}()

	return group, nil
}
