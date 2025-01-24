package message

import (
	"io"
	"log/slog"
)

type InfoMessage struct {
	TrackPriority       TrackPriority
	LatestGroupSequence GroupSequence
	GroupOrder          GroupOrder
}

func (im InfoMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a INFO message")

	/*
	 * Serialize the message in the following format
	 *
	 * INFO Message {
	 *   Message Length (varint),
	 *   Publisher Priority (varint),
	 *   Latest Group Sequence (varint),
	 *   Group Order (varint),
	 *   Group Expires (varint),
	 * }
	 */
	// Serialize the payload
	p := make([]byte, 0, 1<<4)

	// Append the Publisher Priority
	p = appendNumber(p, uint64(im.TrackPriority))

	// Append the Latest Group Sequence
	p = appendNumber(p, uint64(im.LatestGroupSequence))

	// Append the Group Order
	p = appendNumber(p, uint64(im.GroupOrder))

	// Serialize the whole message
	b := make([]byte, 0, len(p)+8)

	// Append the payload
	b = appendBytes(b, p)

	// Write
	_, err := w.Write(b)
	if err != nil {
		slog.Error("failed to write a INFO message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("encoded a INFO message")

	return err
}

func (im *InfoMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a INFO message")

	// Get a message reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get the Publisher Priority
	num, err := readNumber(mr)
	if err != nil {
		return err
	}
	im.TrackPriority = TrackPriority(num)

	// Get the Latest Group Sequence
	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	im.LatestGroupSequence = GroupSequence(num)

	// Get the Group Order
	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	im.GroupOrder = GroupOrder(num)

	slog.Debug("decoded a INFO message")

	return nil
}
