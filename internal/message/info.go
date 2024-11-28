package message

import (
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

type InfoMessage struct {
	PublisherPriority   PublisherPriority
	LatestGroupSequence GroupSequence
	GroupOrder          GroupOrder
	GroupExpires        time.Duration
}

func (im InfoMessage) SerializePayload() []byte {
	/*
	 * Serialize the payload in the following format
	 *
	 * TRACK_STATUS Message {
	 *   Track Namespace (tuple),
	 *   Track Name ([]byte),
	 *   Status Code (varint),
	 *   Last Group ID (varint),
	 *   Last Object ID (varint),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<10)

	// Append the Status Code
	p = quicvarint.Append(p, uint64(im.PublisherPriority))

	// Appen the Last Group ID
	p = quicvarint.Append(p, uint64(im.LatestGroupSequence))

	// Appen the Group Order
	p = quicvarint.Append(p, uint64(im.GroupOrder))

	// Appen the Group Expires
	p = quicvarint.Append(p, uint64(im.GroupExpires))

	return p
}

func (im *InfoMessage) DeserializePayload(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get a Status Code
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	im.PublisherPriority = PublisherPriority(num)

	// Get a Latest Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	im.LatestGroupSequence = GroupSequence(num)

	// Get a Group Order
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	im.GroupOrder = GroupOrder(num)

	// Get a Group Expires
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	im.GroupExpires = time.Duration(num)

	return nil
}
