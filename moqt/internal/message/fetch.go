package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type FetchMessage struct {
	SubscribeID   SubscribeID
	TrackPath     []string
	TrackPriority TrackPriority
	GroupSequence GroupSequence
	FrameSequence FrameSequence // TODO: consider the necessity type FrameSequence
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

	// Append the Subscribe ID
	p = quicvarint.Append(p, uint64(fm.SubscribeID))

	// Append the Track Path Length
	p = quicvarint.Append(p, uint64(len(fm.TrackPath)))

	for _, part := range fm.TrackPath {
		// Append the Track Namespace Prefix Part
		p = quicvarint.Append(p, uint64(len(part)))
		p = append(p, []byte(part)...)
	}

	// Append the Group Priority
	p = quicvarint.Append(p, uint64(fm.TrackPriority))

	// Append the Group Sequence
	p = quicvarint.Append(p, uint64(fm.GroupSequence))

	// Append the Frame Sequence
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

	// Get a Subscribe ID
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	fm.SubscribeID = SubscribeID(num)

	// Get a Track Path Length
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}

	// Get a Track Path
	fm.TrackPath = make([]string, num)

	for i := 0; i < int(num); i++ {
		// Get a Track Path Part
		num, err = quicvarint.Read(mr)
		if err != nil {
			return err
		}

		buf := make([]byte, num)
		_, err = r.Read(buf)
		if err != nil {
			return err
		}

		fm.TrackPath[i] = string(buf)
	}

	// Get a Subscriber Priority
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}

	fm.TrackPriority = TrackPriority(num)

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
