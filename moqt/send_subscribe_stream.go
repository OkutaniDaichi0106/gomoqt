package moqt

// type SendSubscribeStream interface {
// 	// Get the SubscribeID
// 	SubscribeID() SubscribeID

// 	// Get the subscription
// 	SubscribeConfig() SubscribeConfig

// 	// Update the subscription
// 	UpdateSubscribe(SubscribeUpdate) error

// 	// Close the stream
// 	Close() error

// 	// Close the stream with an error
// 	CloseWithError(error) error
// }

// var _ SendSubscribeStream = (*sendSubscribeStream)(nil)

// type sendSubscribeStream struct {
// 	internalStream *internal.SendSubscribeStream
// }

// func (ss *sendSubscribeStream) SubscribeID() SubscribeID {
// 	return SubscribeID(ss.internalStream.SubscribeMessage.SubscribeID)
// }

// func (ss *sendSubscribeStream) SubscribeConfig() SubscribeConfig {
// 	return SubscribeConfig{
// 		TrackPath:        TrackPath(ss.internalStream.SubscribeMessage.TrackPath),
// 		TrackPriority:    TrackPriority(ss.internalStream.SubscribeMessage.TrackPriority),
// 		GroupOrder:       GroupOrder(ss.internalStream.SubscribeMessage.GroupOrder),
// 		MinGroupSequence: GroupSequence(ss.internalStream.SubscribeMessage.MinGroupSequence),
// 		MaxGroupSequence: GroupSequence(ss.internalStream.SubscribeMessage.MaxGroupSequence),
// 	}
// }

// func (ss *sendSubscribeStream) UpdateSubscribe(update SubscribeUpdate) error {
// 	sum := message.SubscribeUpdateMessage{
// 		TrackPriority:    message.TrackPriority(update.TrackPriority),
// 		GroupOrder:       message.GroupOrder(update.GroupOrder),
// 		MinGroupSequence: message.GroupSequence(update.MinGroupSequence),
// 		MaxGroupSequence: message.GroupSequence(update.MaxGroupSequence),
// 	}
// 	return ss.internalStream.SendSubscribeUpdate(sum)
// }

// func (ss *sendSubscribeStream) Close() error {
// 	return ss.internalStream.Close()
// }

// func (ss *sendSubscribeStream) CloseWithError(err error) error {
// 	return ss.internalStream.CloseWithError(err)
// }
