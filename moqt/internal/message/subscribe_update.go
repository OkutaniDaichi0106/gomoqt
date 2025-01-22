package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeUpdateMessage struct {
	TrackPriority    TrackPriority
	GroupOrder       GroupOrder
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence

	SubscribeUpdateParameters Parameters
}

func (su SubscribeUpdateMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a SUBSCRIBE_UPDATE message")
	/*
	 * Serialize the message in the following format
	 *
	 * SUBSCRIBE_UPDATE Message {
	 *   Track Priority (byte),
	 *   Min Group Number (varint),
	 *   Max Group Number (varint),
	 *   Subscribe Update Parameters (Parameters),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<6)

	// Append the Subscriber Priority
	p = quicvarint.Append(p, uint64(su.TrackPriority))

	// Append the Group Order
	p = quicvarint.Append(p, uint64(su.GroupOrder))

	// Append the Min Group Number
	p = quicvarint.Append(p, uint64(su.MinGroupSequence))

	// Append the Max Group Number
	p = quicvarint.Append(p, uint64(su.MaxGroupSequence))

	// Append the Subscribe Update Parameters
	p = appendParameters(p, su.SubscribeUpdateParameters)

	// Get a serialized message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)
	if err != nil {
		slog.Error("failed to write a SUBSCRIBE_UPDATE message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("encoded a SUBSCRIBE_UPDATE message")

	return nil
}

func (sum *SubscribeUpdateMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a SUBSCRIBE_UPDATE message")
	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Track Priority
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	sum.TrackPriority = TrackPriority(num)

	// Get a Group Order
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	sum.GroupOrder = GroupOrder(num)

	// Get a Min Group Number
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	sum.MinGroupSequence = GroupSequence(num)

	// Get a Max Group Number
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	sum.MaxGroupSequence = GroupSequence(num)

	// Get Subscribe Update Parameters
	sum.SubscribeUpdateParameters, err = readParameters(mr)
	if err != nil {
		return err
	}

	slog.Debug("decoded a SUBSCRIBE_UPDATE message")

	return nil
}
