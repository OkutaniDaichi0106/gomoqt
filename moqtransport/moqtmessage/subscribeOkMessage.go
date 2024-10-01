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
	 * Whether contents exist in the track or not
	 */
	ContentExists bool

	/*
	 * Largest Group ID and Largest Object ID available for this track
	 */
	LargestGroupID  GroupID
	LargestObjectID ObjectID

	Parameters Parameters
}

func (so SubscribeOkMessage) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * SUBSCRIBE_UPDATE Message {
	 *   Type (varint) = 0x04,
	 *   Length (varint),
	 *   Subscribe ID (varint),
	 *   Expire (varint),
	 *   Group Order (8),
	 *   Content Exist (flag),
	 *   [Largest Group ID (varint),]
	 *   [Largest Object ID (varint),]
	 *   Parameters
	 * }
	 */

	/*
	 * Serialize the message payload
	 */
	p := make([]byte, 0, 1<<10)

	// Append the Subscriber ID
	p = quicvarint.Append(p, uint64(so.SubscribeID))

	// Append the Expire
	p = quicvarint.Append(p, uint64(so.Expires.Milliseconds()))

	// Append the Group Order
	p = quicvarint.Append(p, uint64(so.GroupOrder))

	// Append the Content Exist
	if so.ContentExists {
		p = quicvarint.Append(p, 1)
		// Append the End Group ID only when the Content Exist is true
		p = quicvarint.Append(p, uint64(so.LargestGroupID))
		// Append the End Object ID only when the Content Exist is true
		p = quicvarint.Append(p, uint64(so.LargestObjectID))
	} else {
		p = quicvarint.Append(p, 0)
	}

	// Append the Parameters
	p = so.Parameters.append(p)

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the type of the message
	b = quicvarint.Append(b, uint64(SUBSCRIBE_OK))

	// Append the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (so *SubscribeOkMessage) DeserializePayload(r quicvarint.Reader) error {
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
