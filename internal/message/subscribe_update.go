package message

import (
	"io"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeUpdateMessage struct {
	SubscribeID SubscribeID

	SubscriberPriority SubscriberPriority
	GroupOrder         GroupOrder
	GroupExpires       time.Duration
	MinGroupSequence   GroupSequence
	MaxGroupSequence   GroupSequence

	Parameters Parameters
}

func (su SubscribeUpdateMessage) SerializePayload(w io.Writer) error {
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
	p := make([]byte, 0, 1<<6)

	// Append the Subscriber ID
	p = quicvarint.Append(p, uint64(su.SubscribeID))

	// Append the Publisher Priority
	p = quicvarint.Append(p, uint64(su.SubscriberPriority))

	// Append the Min Group Number
	p = quicvarint.Append(p, uint64(su.MinGroupSequence))

	// Append the Max Group Number
	p = quicvarint.Append(p, uint64(su.MaxGroupSequence))

	// Append the Subscribe Update Parameters
	p = appendParameters(p, su.Parameters)

	// Get a serialized message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)

	return err
}

func (su *SubscribeUpdateMessage) Decode(r Reader) error {
	// Get a Subscribe ID
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.SubscribeID = SubscribeID(num)

	// Get a Min Group Number
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.MinGroupSequence = GroupSequence(num)

	// Get a Max Group Number
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	su.MaxGroupSequence = GroupSequence(num)

	// Get a Subscriber Priority
	priorityBuf := make([]byte, 1)
	_, err = r.Read(priorityBuf)
	if err != nil {
		return err
	}
	su.SubscriberPriority = SubscriberPriority(priorityBuf[0])

	// Get Subscribe Update Parameters
	su.Parameters, err = readParameters(r)
	if err != nil {
		return err
	}

	return nil
}
