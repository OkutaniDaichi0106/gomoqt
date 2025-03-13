package moqt

import (
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

var _ SendTrackStream = (*sendTrackStream)(nil)

type SendTrackStream interface {
	SubscribeID() SubscribeID
	// SubscribeConfig() SubscribeConfig
	TrackWriter
}

func newSendTrackStream(session *internal.Session, receiveSubscribeStream *internal.ReceiveSubscribeStream) *sendTrackStream {
	return &sendTrackStream{
		session:         session,
		subscribeStream: receiveSubscribeStream,
	}
}

type sendTrackStream struct {
	session             *internal.Session
	subscribeStream     *internal.ReceiveSubscribeStream
	latestGroupSequence GroupSequence
	mu                  sync.RWMutex
}

func (s *sendTrackStream) SubscribeID() SubscribeID {
	return SubscribeID(s.subscribeStream.SubscribeID)
}

func (s *sendTrackStream) TrackPath() TrackPath {
	return TrackPath(s.subscribeStream.TrackPath)
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

	return SubscribeConfig{
		// TrackPath:        s.TrackPath(),
		TrackPriority:    TrackPriority(s.subscribeStream.TrackPriority),
		GroupOrder:       GroupOrder(s.subscribeStream.GroupOrder),
		MinGroupSequence: GroupSequence(s.subscribeStream.MinGroupSequence),
		MaxGroupSequence: GroupSequence(s.subscribeStream.MaxGroupSequence),
	}
}

func (s *sendTrackStream) Info() Info {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Info{
		TrackPriority:       TrackPriority(s.subscribeStream.TrackPriority),
		LatestGroupSequence: s.latestGroupSequence,
		GroupOrder:          GroupOrder(s.subscribeStream.GroupOrder),
	}
}

func (s *sendTrackStream) OpenGroup(sequence GroupSequence) (GroupWriter, error) {
	sgs, err := s.session.OpenGroupStream(message.GroupMessage{
		SubscribeID:   s.subscribeStream.SubscribeID,
		GroupSequence: message.GroupSequence(sequence),
	})
	if err != nil {
		return nil, err
	}

	// Update latest group sequence
	if sequence > s.latestGroupSequence {
		s.latestGroupSequence = sequence
	}

	stream := &sendGroupStream{
		internalStream: sgs,
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
