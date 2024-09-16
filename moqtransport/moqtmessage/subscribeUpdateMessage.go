package moqtmessage

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeUpdateMessage struct {
	/*
	 * A number to identify the subscribe session
	 */
	SubscribeID

	/*
	 * Filter conditions
	 * StartGroupID used only for "AbsoluteStart" or "AbsoluteRange"
	 * StartObjectID used only for "AbsoluteStart" or "AbsoluteRange"
	 * EndGroupID used only for "AbsoluteRange"
	 * EndObjectID used only for "AbsoluteRange". When it is 0, it means the entire group is required
	 */
	FilterRange

	/*
	 * The priority of a subscription relative to other subscriptions in the same session
	 * Lower numbers get higher priority
	 */
	SubscriberPriority

	/*
	 * Subscribe Update Parameters
	 */
	Parameters Parameters
}

func (su SubscribeUpdateMessage) Serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * SUBSCRIBE_UPDATE Message {
	 *   Subscribe ID (varint),
	 *   Start Group ID (varint),
	 *   Start Object ID (varint),
	 *   End Group ID (varint),
	 *   End Object ID (varint),
	 *   Subscriber Priority (8),
	 *   Number of Parameters (varint),
	 *   Subscribe Parameters (..),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(SUBSCRIBE_UPDATE))
	// Append the Subscriber ID
	b = quicvarint.Append(b, uint64(su.SubscribeID))
	// Append the Start Group ID
	b = quicvarint.Append(b, uint64(su.StartGroup))
	// Append the Start Object ID
	b = quicvarint.Append(b, uint64(su.StartObject))
	// Append the End Group ID
	b = quicvarint.Append(b, uint64(su.EndGroup))
	// Append the End Object ID
	b = quicvarint.Append(b, uint64(su.EndObject))
	// Append the Publisher Priority
	b = quicvarint.Append(b, uint64(su.SubscriberPriority))
	// Append the Subscribe Update Priority
	b = su.Parameters.append(b)

	return b
}

// func (su *SubscribeUpdateMessage) deserialize(r quicvarint.Reader) error {
// 	// Get Message ID and check it
// 	id, err := deserializeHeader(r)
// 	if err != nil {
// 		return err
// 	}
// 	if id != SUBSCRIBE_UPDATE {
// 		return errors.New("unexpected message")
// 	}

// 	return su.deserializeBody(r)
// }

func (su *SubscribeUpdateMessage) DeserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.SubscribeID = SubscribeID(num)

	// Get Start Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.StartGroup = GroupID(num)

	// Get Start Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.StartObject = ObjectID(num)

	// Get End Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.EndGroup = GroupID(num)

	// Get End Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.EndObject = ObjectID(num)

	// Get Subscriber Priority
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if num >= 1<<8 {
		return errors.New("publiser priority is not an 8 bit integer")
	}
	su.SubscriberPriority = SubscriberPriority(num)

	// Get Subscribe Update Parameters
	err = su.Parameters.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}
