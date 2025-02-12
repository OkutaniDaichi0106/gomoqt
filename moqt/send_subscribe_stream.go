package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
)

type SendSubscribeStream interface {
	// Get the SubscribeID
	SubscribeID() SubscribeID

	// Get the subscription
	SubscribeConfig() SubscribeConfig

	// Update the subscription
	UpdateSubscribe(SubscribeUpdate) error

	//
	ReceiveSubscribeGap() (SubscribeGap, error)

	// Close the stream
	Close() error

	// Close the stream with an error
	CloseWithError(error) error
}

var _ SendSubscribeStream = (*sendSubscribeStream)(nil)

type sendSubscribeStream struct {
	internalStream *internal.SendSubscribeStream
}

func (ss *sendSubscribeStream) SubscribeID() SubscribeID {
	return SubscribeID(ss.internalStream.SubscribeMessage.SubscribeID)
}

func (ss *sendSubscribeStream) SubscribeConfig() SubscribeConfig {
	return SubscribeConfig{
		TrackPath:           ss.internalStream.SubscribeMessage.TrackPath,
		TrackPriority:       TrackPriority(ss.internalStream.SubscribeMessage.TrackPriority),
		GroupOrder:          GroupOrder(ss.internalStream.SubscribeMessage.GroupOrder),
		MinGroupSequence:    GroupSequence(ss.internalStream.SubscribeMessage.MinGroupSequence),
		MaxGroupSequence:    GroupSequence(ss.internalStream.SubscribeMessage.MaxGroupSequence),
		SubscribeParameters: Parameters{ss.internalStream.SubscribeMessage.SubscribeParameters},
	}
}

func (ss *sendSubscribeStream) UpdateSubscribe(update SubscribeUpdate) error {
	return ss.internalStream.UpdateSubscribe(message.SubscribeUpdateMessage{
		TrackPriority:             message.TrackPriority(update.TrackPriority),
		GroupOrder:                message.GroupOrder(update.GroupOrder),
		MinGroupSequence:          message.GroupSequence(update.MinGroupSequence),
		MaxGroupSequence:          message.GroupSequence(update.MaxGroupSequence),
		SubscribeUpdateParameters: message.Parameters(update.SubscribeParameters.paramMap),
	})
}

func (ss *sendSubscribeStream) ReceiveSubscribeGap() (SubscribeGap, error) {
	gap, err := ss.internalStream.ReceiveSubscribeGap()
	if err != nil {
		return SubscribeGap{}, err
	}

	return SubscribeGap{
		start: GroupSequence(gap.GapStartSequence),
		count: gap.GapCount,
		code:  GroupErrorCode(gap.GroupErrorCode),
	}, nil
}

func (ss *sendSubscribeStream) Close() error {
	return nil //TODO
}

func (ss *sendSubscribeStream) CloseWithError(err error) error {
	return nil //TODO
}
