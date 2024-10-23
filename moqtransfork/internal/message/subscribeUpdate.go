package message

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeUpdateMessage struct {
	SubscribeID SubscribeID

	SubscriberPriority SubscriberPriority

	MinGroupNumber uint64
	MaxGroupNumber uint64

	Parameters Parameters
}

func (su SubscribeUpdateMessage) SerializePayload() []byte {
	/*
	 * Serialize the message in the following format
	 *
	 * SUBSCRIBE_UPDATE Message {
	 *   Subscribe ID (varint),
	 *   Subscriber Priority (byte),
	 *   Min Group Number (varint),
	 *   Max Group Number (varint),
	 *   Subscribe Update Parameters (Parameters),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Subscriber ID
	p = quicvarint.Append(p, uint64(su.SubscribeID))

	// Append the Publisher Priority
	p = quicvarint.Append(p, uint64(su.SubscriberPriority))

	// Append the Min Group Number
	p = quicvarint.Append(p, su.MinGroupNumber)

	// Append the Max Group Number
	p = quicvarint.Append(p, uint64(su.MaxGroupNumber))

	// Append the Subscribe Update Parameters
	p = su.Parameters.Append(p)

	return p
}

func (su *SubscribeUpdateMessage) DeserializePayload(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get a Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.SubscribeID = SubscribeID(num)

	// Get a Min Group Number
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.MinGroupNumber = num

	// Get a Max Group Number
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.MaxGroupNumber = num

	// Get a Subscriber Priority
	priorityBuf := make([]byte, 1)
	_, err = r.Read(priorityBuf)
	if err != nil {
		return err
	}
	su.SubscriberPriority = SubscriberPriority(priorityBuf[0])

	// Get Subscribe Update Parameters
	err = su.Parameters.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}
