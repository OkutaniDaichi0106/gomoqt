package moqt

import (
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

var _ TrackWriter = (*sendTrackStream)(nil)

type sendTrackStream struct {
	session                *internal.Session
	receiveSubscribeStream *internal.ReceiveSubscribeStream
	latestGroupSequence    GroupSequence
	gaps                   map[GroupSequence]struct{}
	mu                     sync.Mutex
}

func newSendTrackStream(session *internal.Session, receiveSubscribeStream *internal.ReceiveSubscribeStream) *sendTrackStream {
	return &sendTrackStream{
		session:                session,
		receiveSubscribeStream: receiveSubscribeStream,
		gaps:                   make(map[GroupSequence]struct{}),
	}
}

func (s *sendTrackStream) TrackPath() TrackPath {
	return TrackPath(s.receiveSubscribeStream.SubscribeMessage.TrackPath)
}

func (s *sendTrackStream) TrackPriority() TrackPriority {
	return TrackPriority(s.receiveSubscribeStream.SubscribeMessage.TrackPriority)
}

func (s *sendTrackStream) GroupOrder() GroupOrder {
	return GroupOrder(s.receiveSubscribeStream.SubscribeMessage.GroupOrder)
}

func (s *sendTrackStream) SubscribeConfig() SubscribeConfig {
	return SubscribeConfig{
		TrackPath:           s.TrackPath(),
		TrackPriority:       s.TrackPriority(),
		GroupOrder:          s.GroupOrder(),
		MinGroupSequence:    GroupSequence(s.receiveSubscribeStream.SubscribeMessage.MinGroupSequence),
		MaxGroupSequence:    GroupSequence(s.receiveSubscribeStream.SubscribeMessage.MaxGroupSequence),
		SubscribeParameters: Parameters{s.receiveSubscribeStream.SubscribeMessage.SubscribeParameters},
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
		SubscribeID:   s.receiveSubscribeStream.SubscribeMessage.SubscribeID,
		GroupSequence: message.GroupSequence(sequence),
	})
	if err != nil {
		return nil, err
	}

	// Update latest group sequence
	if sequence > s.latestGroupSequence {
		for i := s.latestGroupSequence + 1; i < sequence; i++ {
			s.gaps[i] = struct{}{}
		}

		s.latestGroupSequence = sequence
	} else {
		// Remove the gap
		delete(s.gaps, sequence)
	}

	stream := &sendGroupStream{sgs}

	return stream, nil
}

func (s *sendTrackStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.receiveSubscribeStream.Close()
}

func (s *sendTrackStream) CloseWithError(err error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err == nil {
		err = ErrInternalError
	}

	// // Close all group streams
	// max := GroupSequence(s.receiveSubscribeStream.SubscribeMessage.MaxGroupSequence)
	// if s.latestGroupSequence < max {
	// 	var grperr GroupError
	// 	if errors.As(err, &grperr) {
	// 		s.CancelQueue(s.latestGroupSequence+1, uint64(max-s.latestGroupSequence), grperr.GroupErrorCode())
	// 	} else {
	// 		errors.As(ErrInternalError, &grperr)
	// 		s.CancelQueue(s.latestGroupSequence+1, uint64(max-s.latestGroupSequence), grperr.GroupErrorCode())
	// 	}
	// }

	return s.receiveSubscribeStream.CloseWithError(err)
}
