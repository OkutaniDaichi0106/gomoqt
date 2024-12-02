package message

import (
	"io"
	"log/slog"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

type InfoMessage struct {
	PublisherPriority   PublisherPriority
	LatestGroupSequence GroupSequence
	GroupOrder          GroupOrder
	GroupExpires        time.Duration
}

func (im InfoMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a INFO message")

	/*
	 * Serialize the payload in the following format
	 *
	 * TRACK_STATUS Message {
	 *   Track Namespace (tuple),
	 *   Track Name ([]byte),
	 *   Status Code (varint),
	 *   Last Group ID (varint),
	 *   Last Object ID (varint),
	 * }
	 */

	p := make([]byte, 0, 1<<10)

	// Append the Status Code
	p = quicvarint.Append(p, uint64(im.PublisherPriority))

	// Appen the Last Group ID
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

	return err
}

func (im *InfoMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a INFO message")

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Status Code
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	im.PublisherPriority = PublisherPriority(num)

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

	return nil
}
