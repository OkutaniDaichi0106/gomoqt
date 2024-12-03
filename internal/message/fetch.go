package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type FetchMessage struct {
	TrackNamespace     string
	TrackName          string
	SubscriberPriority SubscriberPriority
	GroupSequence      GroupSequence
	FrameSequence      uint64 // TODO: consider the necessity type FrameSequence
}

func (fm FetchMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a FETCH message")

	/*
	 * Serialize the message in the following format
	 *
	 * FETCH Message Payload {
	 *   Track Namespace (string),
	 *   Track Name (string),
	 *   Subscriber Priority (varint),
	 *   Group Sequence (varint),
	 *   Frame Sequence (varint),
	 * }
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Track Namespace
	p = quicvarint.Append(p, uint64(len(fm.TrackNamespace)))
	p = append(p, []byte(fm.TrackNamespace)...)

	// Append the Track Name
	p = quicvarint.Append(p, uint64(len(fm.TrackName)))
	p = append(p, []byte(fm.TrackName)...)

	// Append the Subscriber Priority
	p = quicvarint.Append(p, uint64(fm.SubscriberPriority))

	// Append the Group Sequence
	p = quicvarint.Append(p, uint64(fm.GroupSequence))

	// Append the Group Offset
	p = quicvarint.Append(p, fm.FrameSequence)

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

func (fm *FetchMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a FETCH message")

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Track Namespace
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	fm.TrackNamespace = string(buf)

	// Get a Track Name
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	buf = make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	fm.TrackName = string(buf)

	// Get a Subscriber Priority
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}

	fm.SubscriberPriority = SubscriberPriority(num)

	// Get a Group Sequence
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}

	fm.GroupSequence = GroupSequence(num)

	// Get a Group Offset
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}

	fm.FrameSequence = num

	return nil
}
