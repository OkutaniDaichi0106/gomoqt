package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type FetchMessage struct {
	TrackPath          string
	SubscriberPriority Priority
	GroupSequence      GroupSequence
	FrameSequence      FrameSequence // TODO: consider the necessity type FrameSequence
}

func (fm FetchMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a FETCH message")

	/*
	 * Serialize the message in the following format
	 *
	 * FETCH Message Payload {
	 *   Track Path (string),
	 *   Subscriber Priority (varint),
	 *   Group Sequence (varint),
	 *   Frame Sequence (varint),
	 * }
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Track Path
	p = quicvarint.Append(p, uint64(len(fm.TrackPath)))
	p = append(p, []byte(fm.TrackPath)...)

	// Append the Subscriber Priority
	p = quicvarint.Append(p, uint64(fm.SubscriberPriority))

	// Append the Group Sequence
	p = quicvarint.Append(p, uint64(fm.GroupSequence))

	// Append the Group Offset
	p = quicvarint.Append(p, uint64(fm.FrameSequence))

	// Get a serialized message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)
	if err != nil {
		return err
	}

	slog.Debug("encoded a FETCH message")

	return nil
}

func (fm *FetchMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a FETCH message")

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Track Name
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	fm.TrackPath = string(buf)

	// Get a Subscriber Priority
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}

	fm.SubscriberPriority = Priority(num)

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

	fm.FrameSequence = FrameSequence(num)

	slog.Debug("decoded a FETCH message")

	return nil
}
