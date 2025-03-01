package moqt

// type ReceiveSubscribeStream interface {
// 	SubscribeID() SubscribeID

// 	SubscribeConfig() SubscribeConfig

// 	CloseWithError(error) error
// 	Close() error
// }

// var _ ReceiveSubscribeStream = (*receiveSubscribeStream)(nil)

// type receiveSubscribeStream struct {
// 	internalStream *internal.ReceiveSubscribeStream
// }

// func (rs *receiveSubscribeStream) SubscribeID() SubscribeID {
// 	return SubscribeID(rs.internalStream.SubscribeMessage.SubscribeID)
// }

// func (rs *receiveSubscribeStream) SubscribeConfig() SubscribeConfig {
// 	return SubscribeConfig{
// 		TrackPath:        TrackPath(rs.internalStream.SubscribeMessage.TrackPath),
// 		TrackPriority:    TrackPriority(rs.internalStream.SubscribeMessage.TrackPriority),
// 		GroupOrder:       GroupOrder(rs.internalStream.SubscribeMessage.GroupOrder),
// 		MinGroupSequence: GroupSequence(rs.internalStream.SubscribeMessage.MinGroupSequence),
// 		MaxGroupSequence: GroupSequence(rs.internalStream.SubscribeMessage.MaxGroupSequence),
// 	}
// }

// func (rs *receiveSubscribeStream) Close() error {
// 	return rs.internalStream.Close()
// }

// func (rs *receiveSubscribeStream) CloseWithError(err error) error {
// 	return rs.internalStream.CloseWithError(err)
// }
