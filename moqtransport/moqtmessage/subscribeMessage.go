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
type FilterCode uint64

const (
	LATEST_GROUP   FilterCode = 0x01
	LATEST_OBJECT  FilterCode = 0x02
	ABSOLUTE_START FilterCode = 0x03
	ABSOLUTE_RANGE FilterCode = 0x04
)

type SubscriptionFilter struct {
	/*
	 * Filter FilterCode indicates the type of filter
	 * This indicates whether the StartGroup/StartObject and EndGroup/EndObject fields
	 * will be present
	 */
	FilterCode

	/*
	 * Range of the Filter
	 */
	FilterRange
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

func (sf SubscriptionFilter) isOK() error { //TODO
	switch sf.FilterCode {
	case LATEST_GROUP, LATEST_OBJECT, ABSOLUTE_START:
		return nil
	case ABSOLUTE_RANGE:
		// Check if the Start Group ID is smaller than End Group ID
		if sf.StartGroup > sf.EndGroup {
			return ErrInvalidFilter
		}
		return nil
	default:
		return ErrInvalidFilter
	}
	//TODO: Check if the Filter Code is valid and valid parameters is set
}

func (sf SubscriptionFilter) append(b []byte) []byte {
	if sf.FilterCode == LATEST_GROUP {
		b = quicvarint.Append(b, uint64(sf.FilterCode))
	} else if sf.FilterCode == LATEST_OBJECT {
		b = quicvarint.Append(b, uint64(sf.FilterCode))
	} else if sf.FilterCode == ABSOLUTE_START {
		// Append Filter Type, Start Group ID and Start Object ID
		b = quicvarint.Append(b, uint64(sf.FilterCode))
		b = quicvarint.Append(b, uint64(sf.StartGroup))
		b = quicvarint.Append(b, uint64(sf.StartObject))
	} else if sf.FilterCode == ABSOLUTE_RANGE {
		// Append Filter Type, Start Group ID, Start Object ID, End Group ID and End Object ID
		b = quicvarint.Append(b, uint64(sf.FilterCode))
		b = quicvarint.Append(b, uint64(sf.StartGroup))
		b = quicvarint.Append(b, uint64(sf.StartObject))
		b = quicvarint.Append(b, uint64(sf.EndGroup))
		b = quicvarint.Append(b, uint64(sf.EndObject))
	} else {
		panic("invalid filter")
	}
	return b
}

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

func (s SubscribeMessage) Serialize() []byte {
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

func (s *SubscribeMessage) DeserializeBody(r quicvarint.Reader) error {
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
	s.FilterCode = FilterCode(num)

	switch s.FilterCode {
	case LATEST_GROUP, LATEST_OBJECT:
		//Skip
	case ABSOLUTE_START:
		// Get Start Group ID
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		s.SubscriptionFilter.StartGroup = GroupID(num)

		// Get Start Object ID
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		s.SubscriptionFilter.StartObject = ObjectID(num)
	case ABSOLUTE_RANGE:
		// Get Start Group ID
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		s.SubscriptionFilter.StartGroup = GroupID(num)

		// Get Start Object ID
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		s.SubscriptionFilter.StartObject = ObjectID(num)

		// Get End Group ID
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		s.SubscriptionFilter.EndGroup = GroupID(num)

		// Get End Object ID
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		s.SubscriptionFilter.EndObject = ObjectID(num)
	}

	// Get Subscribe Update Parameters
	err = s.Parameters.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}

var ErrInvalidFilter = errors.New("invalid filter type")
