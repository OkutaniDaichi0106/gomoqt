package moqt

import (
	"context"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

var _ TrackReader = (*receiveTrackStream)(nil)

func newReceiveTrackStream(session *internal.Session, info Info, subscribeStream *internal.SendSubscribeStream) *receiveTrackStream {
	rts := &receiveTrackStream{
		session:             session,
		subscribeStream:     subscribeStream,
		latestGroupSequence: GroupSequence(info.LatestGroupSequence),
		gaps:                make(chan *SubscribeGap, 1),
	} // TODO: Handle the info

	// Receive the subscribe gap
	go func() {
		for {
			sgm, err := subscribeStream.ReceiveSubscribeGap()
			if err != nil {
				return
			}

		}

	}()

	return rts
}

type receiveTrackStream struct {
	session             *internal.Session
	subscribeStream     *internal.SendSubscribeStream
	latestGroupSequence GroupSequence
}

func (s *receiveTrackStream) SubscribeID() SubscribeID {
	return SubscribeID(s.subscribeStream.SubscribeMessage.SubscribeID)
}

func (s *receiveTrackStream) TrackPath() TrackPath {
	return TrackPath(s.subscribeStream.SubscribeMessage.TrackPath)
}

func (s *receiveTrackStream) TrackPriority() TrackPriority {
	return TrackPriority(s.subscribeStream.SubscribeMessage.TrackPriority)
}

func (s *receiveTrackStream) Info() Info {
	return Info{
		TrackPriority:       s.TrackPriority(),
		LatestGroupSequence: s.latestGroupSequence,
		GroupOrder:          s.GroupOrder(),
	}
}

func (s *receiveTrackStream) GroupOrder() GroupOrder {
	return GroupOrder(s.subscribeStream.SubscribeMessage.GroupOrder)
}

func (s *receiveTrackStream) SubscribeConfig() SubscribeConfig {
	return SubscribeConfig{
		TrackPath:        s.subscribeStream.SubscribeMessage.TrackPath,
		TrackPriority:    TrackPriority(s.subscribeStream.SubscribeMessage.TrackPriority),
		GroupOrder:       GroupOrder(s.subscribeStream.SubscribeMessage.GroupOrder),
		MinGroupSequence: GroupSequence(s.subscribeStream.SubscribeMessage.MinGroupSequence),
		MaxGroupSequence: GroupSequence(s.subscribeStream.SubscribeMessage.MaxGroupSequence),
		SubscribeParameters: Parameters{
			paramMap: s.subscribeStream.SubscribeMessage.SubscribeParameters,
		},
	}
}

func (s *receiveTrackStream) Close() error { // TODO: implement
	return s.subscribeStream.Close()
}

func (s *receiveTrackStream) CloseWithError(err error) error { // TODO: implement
	return s.subscribeStream.CloseWithError(err)
}

func (s *receiveTrackStream) AcceptGroup(ctx context.Context) (GroupReader, error) {
	// Check for any gaps first
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:

	}

	// If no gaps were found, proceed with accepting the group stream
	rgs, err := s.session.AcceptGroupStream(ctx, s.subscribeStream.SubscribeMessage.SubscribeID)
	if err != nil {
		return nil, err
	}

	return &receiveGroupStream{rgs}, nil
}

func (s *receiveTrackStream) UpdateSubscribe(update SubscribeUpdate) error {
	sum := message.SubscribeUpdateMessage{
		GroupOrder:                message.GroupOrder(update.GroupOrder),
		TrackPriority:             message.TrackPriority(update.TrackPriority),
		MinGroupSequence:          message.GroupSequence(update.MinGroupSequence),
		MaxGroupSequence:          message.GroupSequence(update.MaxGroupSequence),
		SubscribeUpdateParameters: update.SubscribeParameters.paramMap,
	}

	return s.subscribeStream.UpdateSubscribe(sum)
}
