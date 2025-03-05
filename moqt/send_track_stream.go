package moqt

import (
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

var _ SendTrackStream = (*sendTrackStream)(nil)

type SendTrackStream interface {
	SubscribeID() SubscribeID
	SubscribeConfig() SubscribeConfig
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
	mu                  sync.Mutex
}

func (s *sendTrackStream) SubscribeID() SubscribeID {
	return SubscribeID(s.subscribeStream.SubscribeMessage.SubscribeID)
}

func (s *sendTrackStream) TrackPath() TrackPath {
	return TrackPath(s.subscribeStream.SubscribeMessage.TrackPath)
}

func (s *sendTrackStream) TrackPriority() TrackPriority {
	return TrackPriority(s.subscribeStream.SubscribeMessage.TrackPriority)
}

func (s *sendTrackStream) GroupOrder() GroupOrder {
	return GroupOrder(s.subscribeStream.SubscribeMessage.GroupOrder)
}
func (s *sendTrackStream) LatestGroupSequence() GroupSequence {
	return s.latestGroupSequence
}

func (s *sendTrackStream) SubscribeConfig() SubscribeConfig {
	s.mu.Lock()
	defer s.mu.Unlock()

	return SubscribeConfig{
		TrackPath:        s.TrackPath(),
		TrackPriority:    s.TrackPriority(),
		GroupOrder:       s.GroupOrder(),
		MinGroupSequence: GroupSequence(s.subscribeStream.SubscribeMessage.MinGroupSequence),
		MaxGroupSequence: GroupSequence(s.subscribeStream.SubscribeMessage.MaxGroupSequence),
	}
}

func (s *sendTrackStream) Info() Info {
	return Info{
		TrackPriority:       s.TrackPriority(),
		LatestGroupSequence: s.latestGroupSequence,
		GroupOrder:          s.GroupOrder(),
	}
}

func (s *sendTrackStream) OpenGroup(sequence GroupSequence) (GroupWriter, error) {
	sgs, err := s.session.OpenGroupStream(message.GroupMessage{
		SubscribeID:   s.subscribeStream.SubscribeMessage.SubscribeID,
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
