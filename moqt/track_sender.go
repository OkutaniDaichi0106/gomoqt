package moqt

import (
	"errors"
	"sync"
)

func newTrackSender(substr *receiveSubscribeStream, openGroupFunc func(GroupSequence) (*sendGroupStream, error)) *trackSender {
	track := &trackSender{
		queue:           make(map[*sendGroupStream]struct{}),
		subscribeStream: substr,
		openGroupFunc:   openGroupFunc,
	}

	go func() {
		suberr := <-substr.subscribeCanceledCh
		if suberr != nil {
			track.CloseWithError(suberr.SubscribeErrorCode())
		} else {
			track.Close()
		}
	}()
	return track
}

var _ TrackWriter = (*trackSender)(nil)

type trackSender struct {
	subscribeStream *receiveSubscribeStream

	mu    sync.Mutex
	queue map[*sendGroupStream]struct{}

	openGroupFunc func(GroupSequence) (*sendGroupStream, error)
}

func (s *trackSender) OpenGroup(seq GroupSequence) (GroupWriter, error) {
	if seq == 0 {
		return nil, errors.New("group sequence must not be zero")
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

	group, err := s.openGroupFunc(seq)
	if err != nil {
		return nil, err
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
		stream.CloseWithError(SubscribeCanceledErrorCode)
	}
	s.queue = nil

	return s.subscribeStream.closeWithError(code)
}
