package message

import (
	"io"
	"log/slog"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

type InfoMessage struct {
	TrackPriority       TrackPriority
	LatestGroupSequence GroupSequence
	GroupOrder          GroupOrder
	GroupExpires        time.Duration
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
	p := make([]byte, 0, 1<<10)

	// Append the Publisher Priority
	p = quicvarint.Append(p, uint64(im.TrackPriority))

	// Appen the Latest Group Sequence
	p = quicvarint.Append(p, uint64(im.LatestGroupSequence))

	// Appen the Group Order
	p = quicvarint.Append(p, uint64(im.GroupOrder))

	// Appen the Group Expires
	p = quicvarint.Append(p, uint64(im.GroupExpires))

	// Serialize the whole message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

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

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Publisher Priority
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	im.TrackPriority = TrackPriority(num)

	// Get a Latest Group ID
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	im.LatestGroupSequence = GroupSequence(num)

	// Get a Group Order
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	im.GroupOrder = GroupOrder(num)

	// Get a Group Expires
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	im.GroupExpires = time.Duration(num)

	slog.Debug("decoded a INFO message")

	return nil
}
