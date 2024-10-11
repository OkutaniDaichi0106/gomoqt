package moqtmessage

import "github.com/quic-go/quic-go/quicvarint"

type GroupMessage struct {
	SubscribeID SubscribeID

	GroupID GroupID

	PublisherPriority PublisherPriority
}

func (g GroupMessage) Serialize() []byte {
	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<4)

	// Append the Subscribe ID
	p = quicvarint.Append(p, uint64(g.SubscribeID))

	// Append the Subscribe ID
	p = quicvarint.Append(p, uint64(g.GroupID))

	// Append the Publisher Priority
	p = quicvarint.Append(p, uint64(g.PublisherPriority))

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the type of the message
	b = quicvarint.Append(b, uint64(CLIENT_SETUP))

	// Appen the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
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
	g.GroupID = GroupID(num)

	// Get a Publisher Priority
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	g.PublisherPriority = PublisherPriority(num)

	return nil
}
