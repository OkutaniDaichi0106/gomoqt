package moqtmessage

import (
	"errors"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

type GroupOrder byte

const (
	ASCENDING  GroupOrder = 0x1
	DESCENDING GroupOrder = 0x2
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

	Parameters Parameters
}

func (so SubscribeOkMessage) Serialize() []byte {
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

// func (so *SubscribeOkMessage) deserialize(r quicvarint.Reader) error {
// 	// Get Message ID and check it
// 	id, err := deserializeHeader(r)
// 	if err != nil {
// 		return err
// 	}
// 	if id != SUBSCRIBE_OK {
// 		return errors.New("unexpected message")
// 	}

// 	return so.deserializeBody(r)
// }

func (so *SubscribeOkMessage) DeserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

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
