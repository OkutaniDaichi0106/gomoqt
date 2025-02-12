package moqt

import (
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

type ReceiveSubscribeStream interface {
	SubscribeID() SubscribeID

	SubscribeConfig() SubscribeConfig

	SendSubscribeGap(SubscribeGap) error

	CloseWithError(error) error
	Close() error
}

var _ ReceiveSubscribeStream = (*receiveSubscribeStream)(nil)

type receiveSubscribeStream struct {
	internalStream *internal.ReceiveSubscribeStream
}

func (rs *receiveSubscribeStream) SubscribeID() SubscribeID {
	return SubscribeID(rs.internalStream.SubscribeMessage.SubscribeID)
}

func (rs *receiveSubscribeStream) SubscribeConfig() SubscribeConfig {
	return SubscribeConfig{
		TrackPath:           rs.internalStream.SubscribeMessage.TrackPath,
		TrackPriority:       TrackPriority(rs.internalStream.SubscribeMessage.TrackPriority),
		GroupOrder:          GroupOrder(rs.internalStream.SubscribeMessage.GroupOrder),
		MinGroupSequence:    GroupSequence(rs.internalStream.SubscribeMessage.MinGroupSequence),
		MaxGroupSequence:    GroupSequence(rs.internalStream.SubscribeMessage.MaxGroupSequence),
		SubscribeParameters: Parameters{rs.internalStream.SubscribeMessage.SubscribeParameters},
	}
}

func (rs *receiveSubscribeStream) SendSubscribeGap(gap SubscribeGap) error {
	return rs.internalStream.SendSubscribeGap(message.SubscribeGapMessage{
		GapStartSequence: message.GroupSequence(gap.start),
		GapCount:         gap.count,
		GroupErrorCode:   protocol.GroupErrorCode(gap.code),
	})
}

func (rs *receiveSubscribeStream) Close() error {
	return rs.internalStream.Close()
}

func (rs *receiveSubscribeStream) CloseWithError(err error) error {
	return rs.internalStream.CloseWithError(err)
}
