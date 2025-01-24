package message

import (
	"io"
	"log/slog"
)

type GroupSequence uint64

type GroupMessage struct {
	SubscribeID   SubscribeID
	GroupSequence GroupSequence
	TrackPriority TrackPriority
}

func (g GroupMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a GROUP message")

	/*
	 * Serialize the payload in the following format
	 *
	 * GROUP Message Payload {
	 *   Subscribe ID (varint),
	 *   Group Sequence (varint),
	 *   Publisher Priority (varint),
	 * }
	 */
	p := make([]byte, 0, 1<<4)

	// Append the Subscribe ID
	p = appendNumber(p, uint64(g.SubscribeID))

	// Append the Group Sequence
	p = appendNumber(p, uint64(g.GroupSequence))

	// Append the Publisher Priority
	p = appendNumber(p, uint64(g.TrackPriority))

	// Get a serialized message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload and the payload
	b = appendBytes(b, p)

	// Write the serialized message to the writer
	_, err := w.Write(b)
	if err != nil {
		slog.Error("failed to write a GROUP message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("encoded a GROUP message")

	return nil
}

func (g *GroupMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a GROUP message")

	// Get a message reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get the Subscribe ID
	num, err := readNumber(mr)
	if err != nil {
		return err
	}
	g.SubscribeID = SubscribeID(num)

	// Get the Group Sequence
	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	g.GroupSequence = GroupSequence(num)

	// Get the Publisher Priority
	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	g.TrackPriority = TrackPriority(num)

	slog.Debug("decoded a GROUP message")

	return nil
}
