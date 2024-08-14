package gomoq

import (
	"errors"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeID uint64
type SubscriberPriority byte

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
	StartGroupID  GroupID
	StartObjectID ObjectID
	EndGroupID    GroupID
	EndObjectID   ObjectID

	/*
	 * The priority of a subscription relative to other subscriptions in the same session
	 * Lower numbers get higher priority
	 */
	SubscriberPriority SubscriberPriority

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
	b = quicvarint.Append(b, uint64(su.SubscribeID))
	// Append the Start Group ID
	b = quicvarint.Append(b, uint64(su.StartGroupID))
	// Append the Start Object ID
	b = quicvarint.Append(b, uint64(su.StartObjectID))
	// Append the End Group ID
	b = quicvarint.Append(b, uint64(su.EndGroupID))
	// Append the End Object ID
	b = quicvarint.Append(b, uint64(su.EndObjectID))
	// Append the Publisher Priority
	b = quicvarint.Append(b, uint64(su.SubscriberPriority))
	// Append the Subscribe Update Priority
	b = su.Parameters.append(b)

	return b
}

func (su *SubscribeUpdateMessage) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != SUBSCRIBE_UPDATE {
		return errors.New("unexpected message")
	}

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
	su.StartGroupID = GroupID(num)

	// Get Start Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.StartObjectID = ObjectID(num)

	// Get End Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.EndGroupID = GroupID(num)

	// Get End Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.EndObjectID = ObjectID(num)

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
	err = su.Parameters.parse(r)
	if err != nil {
		return err
	}

	return nil
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

	/*
	 * The type of filter
	 * This indicates whether the StartGroup/StartObject and EndGroup/EndObject fields
	 * will be present
	 */
	FilterType SubscriptionFilter

	// Filter conditions
	/*
	 * StartGroupID used only for "AbsoluteStart" or "AbsoluteRange"
	 */
	StartGroupID GroupID

	/*
	 * StartObjectID used only for "AbsoluteStart" or "AbsoluteRange"
	 */
	StartObjectID ObjectID

	/*
	 * EndGroupID used only for "AbsoluteRange"
	 */
	EndGroupID GroupID

	/*
	 * EndObjectID used only for "AbsoluteRange".
	 * When it is 0, it means the entire group is required
	 */
	EndObjectID ObjectID

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

	if s.FilterType == LATEST_GROUP {
		b = quicvarint.Append(b, uint64(s.FilterType))
	} else if s.FilterType == LATEST_OBJECT {
		b = quicvarint.Append(b, uint64(s.FilterType))
	} else if s.FilterType == ABSOLUTE_START {
		// Append Filter Type, Start Group ID and Start Object ID
		b = quicvarint.Append(b, uint64(s.FilterType))
		b = quicvarint.Append(b, uint64(s.StartGroupID))
		b = quicvarint.Append(b, uint64(s.StartObjectID))
	} else if s.FilterType == ABSOLUTE_RANGE {
		// Append Filter Type, Start Group ID, Start Object ID, End Group ID and End Object ID
		b = quicvarint.Append(b, uint64(s.FilterType))
		b = quicvarint.Append(b, uint64(s.StartGroupID))
		b = quicvarint.Append(b, uint64(s.StartObjectID))
		b = quicvarint.Append(b, uint64(s.EndGroupID))
		b = quicvarint.Append(b, uint64(s.EndObjectID))
	} else {
		panic("invalid filter")
	}

	// Append the Subscribe Update Priority
	b = s.Parameters.append(b)

	return b
}

func (s *SubscribeMessage) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != SUBSCRIBE {
		return errors.New("unexpected message")
	}

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
	s.FilterType = SubscriptionFilter(num)

	// Get Start Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.StartGroupID = GroupID(num)

	// Get Start Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.StartObjectID = ObjectID(num)

	// Get End Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.EndGroupID = GroupID(num)

	// Get End Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	s.EndObjectID = ObjectID(num)

	// Get Subscribe Update Parameters
	err = s.Parameters.parse(r)
	if err != nil {
		return err
	}

	return nil
}

type GroupOrder byte

const (
	NOT_SPECIFY GroupOrder = 0x0
	ASCENDING   GroupOrder = 0x1
	DESCENDING  GroupOrder = 0x2
)

type SubscribeOkMessage struct {
	/*
	 * A number to identify the subscribe session
	 */
	SubscribeID

	/*
	 * Duration in which the permission for subscription is valid
	 * The permission will expire after subscriber receives this message
	 * and the duration has end
	 */
	Expires time.Duration

	/*
	 * The order of the group
	 * This defines how the media is played
	 */
	GroupOrder

	/*
	 * A flag indicating if the specifyed contents
	 */
	ContentExists   bool
	LargestGroupID  GroupID
	LargestObjectID ObjectID
}

func (so SubscribeOkMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * SUBSCRIBE_UPDATE Message {
	 *   Subscribe ID (varint),
	 *   Expire (varint),
	 *   Group Order (8),
	 *   Content Exist (flag),
	 *   [Largest Group ID (varint),]
	 *   [Largest Object ID (varint),]
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(SUBSCRIBE_OK))
	// Append Subscriber ID
	b = quicvarint.Append(b, uint64(so.SubscribeID))
	// Append Expire
	if so.Expires.Microseconds() < 0 {
		so.Expires = 0
	}
	b = quicvarint.Append(b, uint64(so.Expires.Microseconds()))
	// Append Group Order
	b = quicvarint.Append(b, uint64(so.GroupOrder))

	// Append Content Exist
	if !so.ContentExists {
		b = quicvarint.Append(b, 0)
		return b
	} else if so.ContentExists {
		b = quicvarint.Append(b, 1)
		// Append the End Group ID only when the Content Exist is true
		b = quicvarint.Append(b, uint64(so.LargestGroupID))
		// Append the End Object ID only when the Content Exist is true
		b = quicvarint.Append(b, uint64(so.LargestObjectID))
	}

	return b
}

func (so *SubscribeOkMessage) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != SUBSCRIBE_OK {
		return errors.New("unexpected message")
	}

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	so.SubscribeID = SubscribeID(num)

	// Get Expire
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	so.Expires = time.Duration(num)

	// Get Group Order
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if num >= 1<<8 {
		return errors.New("detected group order is not an 8 bit integer")
	}
	so.GroupOrder = GroupOrder(num)

	// Get Content Exist
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if num == 0 {
		so.ContentExists = false
		return nil
	} else if num == 1 {
		so.ContentExists = true
	} else {
		// TODO: terminate the session with a Protocol Violation
		return errors.New("detected value is not flag which takes 0 or 1")
	}

	so.ContentExists = true

	// Get Largest Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	so.LargestGroupID = GroupID(num)

	// Get Largest Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	so.LargestObjectID = ObjectID(num)

	return nil
}

type UnsubscribeMessage struct {
	/*
	 * A number to identify the subscribe session
	 */
	SubscribeID
}

func (us UnsubscribeMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * UNSUBSCRIBE Message {
	 *   Subscribe ID (varint),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(UNSUBSCRIBE))

	// Append Subscirbe ID
	b = quicvarint.Append(b, uint64(us.SubscribeID))

	return b
}

func (us *UnsubscribeMessage) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != UNSUBSCRIBE {
		return errors.New("unexpected message")
	}

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	us.SubscribeID = SubscribeID(num)

	return nil
}

type SubscribeDoneMessage struct {
	/*
	 * A number to identify the subscribe session
	 */
	SubscribeID
	StatusCode    SubscribeDoneStatusCode
	Reason        string
	ContentExists bool
	FinalGroupID  GroupID
	FinalObjectID ObjectID
}

func (sd SubscribeDoneMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * SUBSCRIBE_DONE Message {
	 *   Subscribe ID (varint),
	 *   Status Code (varint),
	 *   Reason ([]byte),
	 *   Content Exist (flag),
	 *   Final Group ID (varint),
	 *   Final Object ID (varint),
	 * }
	 */

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(SUBSCRIBE_DONE))

	// Append Subscirbe ID
	b = quicvarint.Append(b, uint64(sd.SubscribeID))

	// Append Status Code
	b = quicvarint.Append(b, uint64(sd.StatusCode))

	// Append Reason
	b = quicvarint.Append(b, uint64(len(sd.Reason)))
	b = append(b, []byte(sd.Reason)...)

	// Append Content Exist
	if !sd.ContentExists {
		b = quicvarint.Append(b, 0)
		return b
	} else if sd.ContentExists {
		b = quicvarint.Append(b, 1)
		// Append the End Group ID only when the Content Exist is true
		b = quicvarint.Append(b, uint64(sd.FinalGroupID))
		// Append the End Object ID only when the Content Exist is true
		b = quicvarint.Append(b, uint64(sd.FinalObjectID))
	}

	return b
}

func (sd *SubscribeDoneMessage) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != SUBSCRIBE_DONE {
		return errors.New("unexpected message")
	}

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	sd.SubscribeID = SubscribeID(num)

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	sd.StatusCode = SubscribeDoneStatusCode(num)

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	sd.Reason = string(buf)

	// Get Content Exist
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if num == 0 {
		sd.ContentExists = false
		return nil
	} else if num == 1 {
		sd.ContentExists = true
	} else {
		// TODO: terminate the session with a Protocol Violation
		return errors.New("detected value is not flag which takes 0 or 1")
	}

	// Get Largest Group ID only when Content Exist is true
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	sd.FinalGroupID = GroupID(num)

	// Get Largest Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	sd.FinalObjectID = ObjectID(num)

	return nil
}
