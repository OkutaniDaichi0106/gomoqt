package moqt

import (
	"sync"
)

var _ SendTrackStream = (*sendTrackStream)(nil)

type SendTrackStream interface {
	SubscribeID() SubscribeID
	TrackWriter
}

func newSendTrackStream(session *session, receiveSubscribeStream *receiveSubscribeStream) *sendTrackStream {
	return &sendTrackStream{
		session:         session,
		subscribeStream: receiveSubscribeStream,
	}
}

type sendTrackStream struct {
	session             *session
	subscribeStream     *receiveSubscribeStream
	latestGroupSequence GroupSequence
	mu                  sync.RWMutex
}

func (s *sendTrackStream) SubscribeID() SubscribeID {
	return s.subscribeStream.id
}

func (s *sendTrackStream) TrackPath() TrackPath {
	return s.subscribeStream.path
}

// func (s *sendTrackStream) TrackPriority() TrackPriority {
// 	return TrackPriority(s.subscribeStream.TrackPriority)
// }

// func (s *sendTrackStream) GroupOrder() GroupOrder {
// 	return GroupOrder(s.subscribeStream.GroupOrder)
// }

func (s *sendTrackStream) LatestGroupSequence() GroupSequence {
	return s.latestGroupSequence
}

func (s *sendTrackStream) SubscribeConfig() SubscribeConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.subscribeStream.config
}

func (s *sendTrackStream) Info() Info {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Info{
		TrackPriority:       s.subscribeStream.config.TrackPriority,
		LatestGroupSequence: s.latestGroupSequence,
		GroupOrder:          s.subscribeStream.config.GroupOrder,
	}
}

func (s *sendTrackStream) OpenGroup(sequence GroupSequence) (GroupWriter, error) {
	stream, err := s.session.openGroupStream(s.subscribeStream.id, sequence)
	if err != nil {
		return nil, err
	}

	// Update latest group sequence
	if sequence > s.latestGroupSequence {
		s.latestGroupSequence = sequence
	}

	return stream, nil
}

func (s *sendTrackStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.subscribeStream.Close()
}

func (s *sendTrackStream) CloseWithError(err error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err == nil {
		err = ErrInternalError
	}

	return s.subscribeStream.CloseWithError(err)
}
