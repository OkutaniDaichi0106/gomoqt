package moqt

import (
	"context"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

var _ ReceiveTrackStream = (*receiveTrackStream)(nil)

type ReceiveTrackStream interface {
	TrackReader
	SubscribeID() SubscribeID
	UpdateSubscribe(*SubscribeConfig) error
}

func newReceiveTrackStream(session *session, info Info, subscribeStream *sendSubscribeStream) *receiveTrackStream {
	rts := &receiveTrackStream{
		session:             session,
		subscribeStream:     subscribeStream,
		latestGroupSequence: GroupSequence(info.LatestGroupSequence),
	}

	// TODO: Handle the info properly, maybe validate or process it further

	return rts
}

type receiveTrackStream struct {
	session             *session
	subscribeStream     *sendSubscribeStream
	latestGroupSequence GroupSequence
}

func (s *receiveTrackStream) SubscribeID() SubscribeID {
	return s.subscribeStream.id
}

func (s *receiveTrackStream) TrackPath() TrackPath {
	return s.subscribeStream.path
}

func (s *receiveTrackStream) TrackPriority() TrackPriority {
	return s.subscribeStream.config.TrackPriority
}

func (s *receiveTrackStream) Info() Info {
	return Info{
		TrackPriority:       s.TrackPriority(),
		LatestGroupSequence: s.latestGroupSequence,
		GroupOrder:          s.GroupOrder(),
	}
}

func (s *receiveTrackStream) GroupOrder() GroupOrder {
	return s.subscribeStream.config.GroupOrder
}

func (s *receiveTrackStream) LatestGroupSequence() GroupSequence {
	return s.latestGroupSequence
}

func (s *receiveTrackStream) Close() error { // TODO: implement
	return s.subscribeStream.Close()
}

func (s *receiveTrackStream) CloseWithError(err error) error { // TODO: implement
	return s.subscribeStream.CloseWithError(err)
}

func (s *receiveTrackStream) AcceptGroup(ctx context.Context) (GroupReader, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:

	}

	rgs, err := s.session.acceptGroupStream(ctx, s.subscribeStream.id)
	if err != nil {
		return nil, err
	}

	return rgs, nil
}

func (s *receiveTrackStream) UpdateSubscribe(update *SubscribeConfig) error {
	if update == nil {
		return nil
	}

	sum := message.SubscribeUpdateMessage{
		GroupOrder:       message.GroupOrder(update.GroupOrder),
		TrackPriority:    message.TrackPriority(update.TrackPriority),
		MinGroupSequence: message.GroupSequence(update.MinGroupSequence),
		MaxGroupSequence: message.GroupSequence(update.MaxGroupSequence),
	}

	return s.subscribeStream.SendSubscribeUpdate(sum)
}
