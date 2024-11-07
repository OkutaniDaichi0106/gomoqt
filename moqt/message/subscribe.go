package message

import (
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeID uint64

type TrackAlias uint64

type SubscriberPriority byte

type GroupOrder byte

type SubscribeMessage struct {
	SubscribeID SubscribeID

	//TrackAlias TrackAlias

	TrackNamespace TrackNamespace

	TrackName string

	SubscriberPriority SubscriberPriority

	/*
	 * The order of the group
	 * This defines how the media is played
	 */
	GroupOrder GroupOrder

	Expires time.Duration

	/***/
	MinGroupSequence GroupSequence

	/***/
	MaxGroupSequence GroupSequence

	/*
	 * Subscribe Parameters
	 * Parameters should contain Authorization Information
	 */
	Parameters Parameters
}

func (s SubscribeMessage) SerializePayload() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * SUBSCRIBE Message payload {
	 *   Track Namespace (Track Namespace),
	 *   Track Name (string),
	 *   Subscriber Priority (8),
	 *   Group Order (8),
	 *   Group Expires (varint),
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
	//p = quicvarint.Append(p, uint64(s.TrackAlias))

	// Append the Track Namespace
	p = AppendTrackNamespace(p, s.TrackNamespace)

	// Append the Track Name
	p = quicvarint.Append(p, uint64(len(s.TrackName)))
	p = append(p, []byte(s.TrackName)...)

	// Append the Subscriber Priority
	p = append(p, []byte{byte(s.SubscriberPriority)}...)

	// Append the Group Order
	p = append(p, []byte{byte(s.GroupOrder)}...)

	// Append the Group Expires
	p = append(p, []byte{byte(s.GroupOrder)}...)

	// Append the Min Group Sequence
	p = quicvarint.Append(p, uint64(s.MinGroupSequence))

	// Append the Max Group Sequence
	p = quicvarint.Append(p, uint64(s.MinGroupSequence))

	// Append the Subscribe Update Priority
	p = s.Parameters.Append(p)

	return p
}

func (s *SubscribeMessage) DeserializePayload(r quicvarint.Reader) error {
	// Get Subscribe ID
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.SubscribeID = SubscribeID(num)

	// // Get Track Alias
	// num, err = quicvarint.Read(r)
	// if err != nil {
	// 	return err
	// }
	// s.TrackAlias = TrackAlias(num)

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
	bnum, err := r.ReadByte()
	if err != nil {
		return err
	}
	s.SubscriberPriority = SubscriberPriority(bnum)

	// Get Group Order
	bnum, err = r.ReadByte()
	if err != nil {
		return err
	}
	s.GroupOrder = GroupOrder(bnum)

	// Get Group Expires
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.Expires = time.Duration(num)

	// Get Min Group Sequence
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.MinGroupSequence = GroupSequence(num)

	// Get Max Group Sequence
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.MaxGroupSequence = GroupSequence(num)

	// Get Subscribe Update Parameters
	err = s.Parameters.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}
