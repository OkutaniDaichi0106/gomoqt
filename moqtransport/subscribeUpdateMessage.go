package moqtransport

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeUpdateMessage struct {
	/*
	 * A number to identify the subscribe session
	 */
	subscribeID

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

func (su SubscribeUpdateMessage) serialize() []byte {
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
	b = quicvarint.Append(b, uint64(su.subscribeID))
	// Append the Start Group ID
	b = quicvarint.Append(b, uint64(su.startGroup))
	// Append the Start Object ID
	b = quicvarint.Append(b, uint64(su.startObject))
	// Append the End Group ID
	b = quicvarint.Append(b, uint64(su.endGroup))
	// Append the End Object ID
	b = quicvarint.Append(b, uint64(su.endObject))
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

func (su *SubscribeUpdateMessage) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.subscribeID = subscribeID(num)

	// Get Start Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.startGroup = groupID(num)

	// Get Start Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.startObject = objectID(num)

	// Get End Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.endGroup = groupID(num)

	// Get End Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.endObject = objectID(num)

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
	err = su.Parameters.deserialize(r)
	if err != nil {
		return err
	}

	return nil
}
