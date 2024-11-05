package message

import "github.com/quic-go/quic-go/quicvarint"

type GroupSequence uint64

type PublisherPriority byte

type GroupMessage struct {
	SubscribeID SubscribeID

	GroupSequence GroupSequence

	PublisherPriority PublisherPriority
}

func (g GroupMessage) SerializePayload() []byte {
	/*
	 * Serialize the payload in the following format
	 *
	 * GROUP Message Payload {
	 *   Subscribe ID (varint),
	 *   Group ID (varint),
	 *   Publisher Priority (varint),
	 * }
	 */
	p := make([]byte, 0, 1<<4)

	// Append the Subscribe ID
	p = quicvarint.Append(p, uint64(g.SubscribeID))

	// Append the Subscribe ID
	p = quicvarint.Append(p, uint64(g.GroupSequence))

	// Append the Publisher Priority
	p = quicvarint.Append(p, uint64(g.PublisherPriority))

	return p
}

func (g *GroupMessage) DeserializePayload(r quicvarint.Reader) error {
	// Get a Subscribe ID
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}
	g.SubscribeID = SubscribeID(num)

	// Get a Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	g.GroupSequence = GroupSequence(num)

	// Get a Publisher Priority
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	g.PublisherPriority = PublisherPriority(num)

	return nil
}
