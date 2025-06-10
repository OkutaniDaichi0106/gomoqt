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
	s.mu.Lock()
	defer s.mu.Unlock()

	err, ok := s.subscribeStream.done()
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
		delete(s.queue, group)
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

	err, ok := s.subscribeStream.done()
	if ok {
		return err
	}

	return s.subscribeStream.close()
}

func (s *trackSender) CloseWithError(code SubscribeErrorCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for stream := range s.queue {
		stream.CloseWithError(SubscribeCanceledErrorCode)
	}
	s.queue = nil

	err, ok := s.subscribeStream.done()
	if ok {
		return err
	}

	return s.subscribeStream.closeWithError(code)
}
