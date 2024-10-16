package moqtmessage

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeID uint64

type TrackAlias uint64

type SubscriberPriority byte

type GroupOrder byte

type SubscribeMessage struct {
	SubscribeID SubscribeID

	TrackAlias TrackAlias

	TrackNamespace TrackNamespace

	TrackName string

	SubscriberPriority SubscriberPriority

	/*
	 * The order of the group
	 * This defines how the media is played
	 */
	GroupOrder GroupOrder

	/***/
	MinGroupSequence uint64

	/***/
	MaxGroupSequence uint64

	/*
	 * Subscribe Parameters
	 * Parameters should contain Authorization Information
	 */
	Parameters Parameters
}

func (s SubscribeMessage) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * SUBSCRIBE Message {
	 *   Type (varint) = 0x03,
	 *   Length (varint),
	 *   Subscribe ID (varint),
	 *   Track Alias (varint),
	 *   Track Namespace (Track Namespace),
	 *   Track Name (string),
	 *   Subscriber Priority (8),
	 *   Group Order (8),
	 *   Min Group Sequence (varint),
	 *   Max Group Sequence (varint),
	 *   Subscribe Parameters (Parameters),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Subscriber ID
	p = quicvarint.Append(p, uint64(s.SubscribeID))

	// Append the Subscriber ID
	p = quicvarint.Append(p, uint64(s.TrackAlias))

	// Append the Track Namespace
	p = AppendTrackNamespace(p, s.TrackNamespace)

	// Append the Track Name
	p = quicvarint.Append(p, uint64(len(s.TrackName)))
	p = append(p, []byte(s.TrackName)...)

	// Append the Subscriber Priority
	p = quicvarint.Append(p, uint64(s.SubscriberPriority))

	// Append the Group Order
	p = quicvarint.Append(p, uint64(s.GroupOrder))

	// Append the Min Group Sequence
	p = quicvarint.Append(p, s.MinGroupSequence)

	// Append the Max Group Sequence
	p = quicvarint.Append(p, s.MinGroupSequence)

	// Append the Subscribe Update Priority
	p = s.Parameters.append(p)

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the type of the message
	b = quicvarint.Append(b, uint64(SUBSCRIBE))

	// Append the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (s *SubscribeMessage) DeserializePayload(r quicvarint.Reader) error {
	// Get Subscribe ID
	num, err := quicvarint.Read(r)
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
	tns, err := ReadTrackNamespace(r)
	if err != nil {
		return err
	}
	s.TrackNamespace = tns

	// Get Track Name
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
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
	s.GroupOrder = GroupOrder(num)

	// Get Min Group Sequence
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.MinGroupSequence = num

	// Get Max Group Sequence
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.MaxGroupSequence = num

	// Get Subscribe Update Parameters
	err = s.Parameters.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}
