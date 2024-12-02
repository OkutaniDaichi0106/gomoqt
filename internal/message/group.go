package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type GroupSequence uint64

type PublisherPriority byte

type GroupMessage struct {
	SubscribeID SubscribeID

	GroupSequence GroupSequence

	PublisherPriority PublisherPriority
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
	p = quicvarint.Append(p, uint64(g.SubscribeID))

	// Append the Subscribe ID
	p = quicvarint.Append(p, uint64(g.GroupSequence))

	// Append the Publisher Priority
	p = quicvarint.Append(p, uint64(g.PublisherPriority))

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

func (g *GroupMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a GROUP message")

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Subscribe ID
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	g.SubscribeID = SubscribeID(num)

	// Get a Subscribe ID
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	g.GroupSequence = GroupSequence(num)

	// Get a Publisher Priority
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	g.PublisherPriority = PublisherPriority(num)

	return nil
}
