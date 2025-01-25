package message

import (
	"bytes"
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type InfoMessage struct {
	TrackPriority       TrackPriority
	LatestGroupSequence GroupSequence
	GroupOrder          GroupOrder
}

func (im InfoMessage) Encode(w io.Writer) (int, error) {
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
	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))

	// Append the payload
	b = appendBytes(b, p)

	// Write
	return w.Write(b)
}

func (im *InfoMessage) Decode(r io.Reader) (int, error) {
	slog.Debug("decoding a INFO message")

	// Read the payload
	buf, n, err := readBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	// Decode the payload
	mr := bytes.NewReader(buf)

	num, _, err := readNumber(mr)
	if err != nil {
		return n, err
	}
	im.TrackPriority = TrackPriority(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	im.LatestGroupSequence = GroupSequence(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	im.GroupOrder = GroupOrder(num)

	slog.Debug("decoded a INFO message")
	return n, nil
}
