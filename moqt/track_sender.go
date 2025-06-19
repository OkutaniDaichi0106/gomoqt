package moqt

import (
	"errors"
	"log/slog"
	"sync"
)

func newTrackSender(substr *receiveSubscribeStream, openGroupFunc func(GroupSequence) (*sendGroupStream, error)) *trackSender {
	track := &trackSender{
		queue:           make(map[*sendGroupStream]struct{}),
		subscribeStream: substr,
		openGroupFunc:   openGroupFunc,
	}

	go func() {
		<-substr.canceled()
		track.mu.Lock()
		defer track.mu.Unlock()
		for stream := range track.queue {
			stream.CancelWrite(SubscribeCanceledErrorCode)
		}
		track.queue = nil
	}()
	return track
}

var _ TrackWriter = (*trackSender)(nil)

type trackSender struct {
	subscribeStream *receiveSubscribeStream

	accepted bool
	info     Info

	mu    sync.Mutex
	queue map[*sendGroupStream]struct{}

	openGroupFunc func(GroupSequence) (*sendGroupStream, error)
}

func (s *trackSender) WriteInfo(info Info) {
	if s.accepted {
		slog.Warn("moq: superfluous accept call on moqt.WriteInfo")
		return
	}

	s.accepted = true
	s.info = info

	s.subscribeStream.accept(info)
}

func (s *trackSender) OpenGroup(seq GroupSequence) (GroupWriter, error) {
	if seq == 0 {
		return nil, errors.New("group sequence must not be zero")
	}

	select {
	case <-s.subscribeStream.subscribeCanceledCh:
		return nil, errors.New("subscription was canceled")
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	err, ok := s.subscribeStream.isClosed()
	if ok {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("track is closed")
	}

	if !s.accepted {
		s.WriteInfo(Info{})
	}

	group, err := s.openGroupFunc(seq)
	if err != nil {
		return nil, err
	}

	if s.queue == nil {
		return nil, errors.New("subscription was canceled")
	}

	s.queue[group] = struct{}{}
	go func() {
		<-group.closedCh
		s.mu.Lock()
		delete(s.queue, group)
		s.mu.Unlock()
	}()

	return group, nil
}

func (s *trackSender) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for stream := range s.queue {
		stream.Close()
	}
	s.queue = nil

	return s.subscribeStream.close()
}

func (s *trackSender) CloseWithError(code SubscribeErrorCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for stream := range s.queue {
		stream.CancelWrite(SubscribeCanceledErrorCode)
	}
	s.queue = nil

	return s.subscribeStream.closeWithError(code)
}
