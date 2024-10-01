package moqtmessage

import (
	"errors"
	"log"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * Subscribe ID
 *
 * An integer that is unique and monotonically increasing within a session
 * and is less than the session's Maximum Subscriber ID
 */
type SubscribeID uint64

/*
 * Priority of a subscription
 *
 * A priority of a subscription relative to other subscriptions in the same session.
 * Lower numbers get higher priority.
 */
type SubscriberPriority byte

/*
 * Filter of the subscription
 *
 * Following type are defined in the official document
 * LATEST_GROUP
 * LATEST_OBJECT
 * ABSOLUTE_START
 * ABSOLUTE_RANGE
 */
type Code uint64

const (
	LATEST_GROUP   Code = 0x01
	LATEST_OBJECT  Code = 0x02
	ABSOLUTE_START Code = 0x03
	ABSOLUTE_RANGE Code = 0x04
)

type SubscriptionFilter struct {
	/*
	 * Filter Code indicates the type of filter
	 * This indicates whether the Range.StartGroup/StartObject and EndGroup/EndObject fields
	 * will be present
	 */
	Code Code

	/*
	 * Range of the Filter
	 */
	Range FilterRange
}

/*
 * Range of the filter
 *
 * This consist of start group ID, start object ID, end group ID and end object ID
 */
type FilterRange struct {
	/*
	 * Startmoqtobject.GroupID used only for "AbsoluteStart" or "AbsoluteRange"
	 */
	StartGroup GroupID

	/*
	 * Startmoqtobject.ObjectID used only for "AbsoluteStart" or "AbsoluteRange"
	 */
	StartObject ObjectID

	/*
	 * Endmoqtobject.GroupID used only for "AbsoluteRange"
	 */
	EndGroup GroupID

	/*
	 * Endmoqtobject.ObjectID used only for "AbsoluteRange".
	 * When it is 0, it means the entire group is required
	 */
	EndObject ObjectID
}

func (sf SubscriptionFilter) append(b []byte) []byte {
	if sf.Code == LATEST_GROUP {
		b = quicvarint.Append(b, uint64(sf.Code))
	} else if sf.Code == LATEST_OBJECT {
		b = quicvarint.Append(b, uint64(sf.Code))
	} else if sf.Code == ABSOLUTE_START {
		// Append the Filter Type, Start Group ID and Start Object ID
		b = quicvarint.Append(b, uint64(sf.Code))
		b = quicvarint.Append(b, uint64(sf.Range.StartGroup))
		b = quicvarint.Append(b, uint64(sf.Range.StartObject))
	} else if sf.Code == ABSOLUTE_RANGE {
		// Append the Filter Type, Start Group ID, Start Object ID, End Group ID and End Object ID
		b = quicvarint.Append(b, uint64(sf.Code))
		b = quicvarint.Append(b, uint64(sf.Range.StartGroup))
		b = quicvarint.Append(b, uint64(sf.Range.StartObject))
		b = quicvarint.Append(b, uint64(sf.Range.EndGroup))
		b = quicvarint.Append(b, uint64(sf.Range.EndObject))
	} else {
		panic("invalid filter")
	}
	return b
}

type SubscribeMessage struct {
	/*
	 * A number to identify the subscribe session
	 */
	SubscribeID        SubscribeID
	TrackAlias         TrackAlias
	TrackNamespace     TrackNamespace
	TrackName          string
	SubscriberPriority SubscriberPriority

	/*
	 * The order of the group
	 * This defines how the media is played
	 */
	GroupOrder GroupOrder

	/***/
	SubscriptionFilter SubscriptionFilter

	/*
	 * Subscribe Parameters
	 * Parameters should include Track Authorization Information
	 */
	Parameters Parameters
}

func (s SubscribeMessage) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * SUBSCRIBE_UPDATE Message {
	 *   Type (varint) = 0x03,
	 *   Length (varint),
	 *   Subscribe ID (varint),
	 *   Track Alias (varint),
	 *   Track Namespace (tuple),
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

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Subscriber ID
	p = quicvarint.Append(p, uint64(s.SubscribeID))

	// Append the Subscriber ID
	p = quicvarint.Append(p, uint64(s.TrackAlias))

	// Append the Track Namespace
	p = s.TrackNamespace.Append(p)

	// Append the Track Name
	p = quicvarint.Append(p, uint64(len(s.TrackName)))
	p = append(p, []byte(s.TrackName)...)

	// Append the Subscriber Priority
	p = quicvarint.Append(p, uint64(s.SubscriberPriority))

	// Append the Group Order
	p = quicvarint.Append(p, uint64(s.GroupOrder))

	// Append the subscription filter
	p = s.SubscriptionFilter.append(p)

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
	var tns TrackNamespace
	err = tns.Deserialize(r)
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
	log.Println("REACH 131", s.TrackName)
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
	s.SubscriptionFilter.Code = Code(num)

	switch s.SubscriptionFilter.Code {
	case LATEST_GROUP, LATEST_OBJECT:
		//Skip
	case ABSOLUTE_START:
		// Get Start Group ID
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		s.SubscriptionFilter.Range.StartGroup = GroupID(num)

		// Get Start Object ID
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		s.SubscriptionFilter.Range.StartObject = ObjectID(num)
	case ABSOLUTE_RANGE:
		// Get Start Group ID
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		s.SubscriptionFilter.Range.StartGroup = GroupID(num)

		// Get Start Object ID
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		s.SubscriptionFilter.Range.StartObject = ObjectID(num)

		// Get End Group ID
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		s.SubscriptionFilter.Range.EndGroup = GroupID(num)

		// Get End Object ID
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		s.SubscriptionFilter.Range.EndObject = ObjectID(num)
	}

	// Get Subscribe Update Parameters
	err = s.Parameters.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}

var ErrInvalidFilter = errors.New("invalid filter type")
