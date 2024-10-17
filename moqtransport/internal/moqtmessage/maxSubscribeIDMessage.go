package moqtmessage

// TODO:

// type MaxSubscribeID struct {
// 	subscribeID
// }

// func (msi MaxSubscribeID) serialize() []byte {
// 	/*
// 	 * Serialize the message in the following formatt
// 	 *
// 	 * MAX_SUBSCRIBE_ID Message {
// 	 *   Max Subscirbe ID (varint),
// 	 * }
// 	 */
// 	b := make([]byte, 0, 1<<4)

// 	b = quicvarint.Append(b, uint64(MAX_SUBSCRIBE_ID))

// 	// Append the Track Namespace Prefix
// 	b = quicvarint.Append(b, uint64(msi.subscribeID))

// 	return b
// }

// func (msi *MaxSubscribeID) deserialize(r quicvarint.Reader) error {
// 	num, err := quicvarint.Read(r)
// 	if err != nil {
// 		return err
// 	}
// 	msi.subscribeID = subscribeID(num)

// 	return err
// }
