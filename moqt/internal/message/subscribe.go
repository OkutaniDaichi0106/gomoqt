package message

import (
	"io"
	"log/slog"
)

type SubscribeID uint64
type TrackPriority byte
type GroupOrder byte

/*
 * SUBSCRIBE Message {
 *   Subscribe ID (varint),
 *   Track Path ([]string),
 *   Track Priority (varint),
 *   Group Order (varint),
 *   Min Group Sequence (varint),
 *   Max Group Sequence (varint),
 *   Subscribe Parameters (Parameters),
 * }
 */
type SubscribeMessage struct {
	SubscribeID         SubscribeID
	TrackPath           []string
	TrackPriority       TrackPriority
	GroupOrder          GroupOrder
	MinGroupSequence    GroupSequence
	MaxGroupSequence    GroupSequence
	SubscribeParameters Parameters
}

func (s SubscribeMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a SUBSCRIBE message")

	// Serialize the payload
	p := make([]byte, 0, 1<<6)
	p = appendNumber(p, uint64(s.SubscribeID))
	p = appendStringArray(p, s.TrackPath)
	p = appendNumber(p, uint64(s.TrackPriority))
	p = appendNumber(p, uint64(s.GroupOrder))
	p = appendNumber(p, uint64(s.MinGroupSequence))
	p = appendNumber(p, uint64(s.MaxGroupSequence))
	p = appendParameters(p, s.SubscribeParameters)

	// Prepare the final message with length prefix
	b := make([]byte, 0, len(p)+8)
	b = appendNumber(b, uint64(len(p)))
	b = append(b, p...)

	// Write the message
	if _, err := w.Write(b); err != nil {
		slog.Error("failed to write a SUBSCRIBE message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("encoded a SUBSCRIBE message")
	return nil
}

func (s *SubscribeMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a SUBSCRIBE message")

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
	s.SubscribeID = SubscribeID(num)

	s.TrackPath, err = readStringArray(mr)
	if err != nil {
		return err
	}

	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	s.TrackPriority = TrackPriority(num)

	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	s.GroupOrder = GroupOrder(num)

	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	s.MinGroupSequence = GroupSequence(num)

	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	s.MaxGroupSequence = GroupSequence(num)

	s.SubscribeParameters, err = readParameters(mr)
	if err != nil {
		return err
	}

	slog.Debug("decoded a SUBSCRIBE message")
	return nil
}
