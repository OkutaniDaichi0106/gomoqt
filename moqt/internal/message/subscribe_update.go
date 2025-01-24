package message

import (
	"io"
	"log/slog"
)

/*
 * SUBSCRIBE_UPDATE Message {
 *   Track Priority (varint),
 *   Group Order (varint),
 *   Min Group Sequence (varint),
 *   Max Group Sequence (varint),
 *   Subscribe Update Parameters (Parameters),
 * }
 */
type SubscribeUpdateMessage struct {
	TrackPriority             TrackPriority
	GroupOrder                GroupOrder
	MinGroupSequence          GroupSequence
	MaxGroupSequence          GroupSequence
	SubscribeUpdateParameters Parameters
}

func (su SubscribeUpdateMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a SUBSCRIBE_UPDATE message")

	// Serialize the payload
	p := make([]byte, 0, 1<<6)
	p = appendNumber(p, uint64(su.TrackPriority))
	p = appendNumber(p, uint64(su.GroupOrder))
	p = appendNumber(p, uint64(su.MinGroupSequence))
	p = appendNumber(p, uint64(su.MaxGroupSequence))
	p = appendParameters(p, su.SubscribeUpdateParameters)

	// Prepare the final message with length prefix
	b := make([]byte, 0, len(p)+8)
	b = appendBytes(b, p)

	// Write the message
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

	// Create a message reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Deserialize the payload
	num, err := readNumber(mr)
	if err != nil {
		return err
	}
	sum.TrackPriority = TrackPriority(num)

	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	sum.GroupOrder = GroupOrder(num)

	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	sum.MinGroupSequence = GroupSequence(num)

	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	sum.MaxGroupSequence = GroupSequence(num)

	sum.SubscribeUpdateParameters, err = readParameters(mr)
	if err != nil {
		return err
	}

	slog.Debug("decoded a SUBSCRIBE_UPDATE message")
	return nil
}
