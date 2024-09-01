package gomoq

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeID uint64
type SubscriberPriority byte

type SubscribeMessage struct {
	/*
	 * A number to identify the subscribe session
	 */
	SubscribeID
	TrackAlias
	TrackNamespace string
	TrackName      string
	SubscriberPriority

	/*
	 * The order of the group
	 * This defines how the media is played
	 */
	GroupOrder GroupOrder

	/***/
	SubscriptionFilter

	/*
	 * Subscribe Parameters
	 * Parameters should include Track Authorization Information
	 */
	Parameters Parameters
}

func (s SubscribeMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * SUBSCRIBE_UPDATE Message {
	 *   Subscribe ID (varint),
	 *   Track Alias (varint),
	 *   Track Namespace ([]byte),
	 *   Track Name ([]byte),
	 *   Subscriber Priority (8),
	 *   Group Order (8),
	 *   Filter Type (varint),
	 *   Start Group ID (varint),
	 *   Start Object ID (varint),
	 *   End Group ID (varint),
	 *   End Object ID (varint),
	 *   Number of Parameters (varint),
	 *   Subscribe Parameters (..),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(SUBSCRIBE))
	// Append Subscriber ID
	b = quicvarint.Append(b, uint64(s.SubscribeID))
	// Append Subscriber ID
	b = quicvarint.Append(b, uint64(s.TrackAlias))
	// Append Track Namespace
	b = quicvarint.Append(b, uint64(len(s.TrackNamespace)))
	b = append(b, []byte(s.TrackNamespace)...)
	// Append Track Name
	b = quicvarint.Append(b, uint64(len(s.TrackName)))
	b = append(b, []byte(s.TrackName)...)
	// Append Subscriber Priority
	b = quicvarint.Append(b, uint64(s.SubscriberPriority))
	// Append Group Order
	b = quicvarint.Append(b, uint64(s.GroupOrder))

	// Append the subscription filter
	b = s.SubscriptionFilter.append(b)

	// Append the Subscribe Update Priority
	b = s.Parameters.append(b)

	return b
}

// func (s *SubscribeMessage) deserialize(r quicvarint.Reader) error {
// 	// Get Message ID and check it
// 	id, err := deserializeHeader(r)
// 	if err != nil {
// 		return err
// 	}
// 	if id != SUBSCRIBE {
// 		return ErrUnexpectedMessage
// 	}

// 	return s.deserializeBody(r)
// }

func (s *SubscribeMessage) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.SubscribeID = SubscribeID(num)

	// Get Track Alias
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.TrackAlias = TrackAlias(num)

	// Get Track Namespace
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	s.TrackNamespace = string(buf)

	// Get Track Name
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	buf = make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	s.TrackName = string(buf)

	// Get Subscriber Priority
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if num >= 1<<8 {
		return errors.New("publiser priority is not an 8 bit integer")
	}
	s.SubscriberPriority = SubscriberPriority(num)

	// Get Group Order
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if num >= 1<<8 {
		return errors.New("publiser priority is not an 8 bit integer")
	}
	s.GroupOrder = GroupOrder(num)

	// Get Filter Type
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.SubscriptionFilter.FilterCode = FilterCode(num)

	// Get Start Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.SubscriptionFilter.startGroup = GroupID(num)

	// Get Start Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.SubscriptionFilter.startObject = ObjectID(num)

	// Get End Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.SubscriptionFilter.endGroup = GroupID(num)

	// Get End Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.SubscriptionFilter.endObject = ObjectID(num)

	// Get Subscribe Update Parameters
	err = s.Parameters.parse(r)
	if err != nil {
		return err
	}

	return nil
}
