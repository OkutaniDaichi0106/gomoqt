package message

import (
	"io"
	"log/slog"
)

/*
 * FETCH Message {
 *   Subscribe ID (varint),
 *   Track Path ([]string),
 *   Track Priority (varint),
 *   Group Sequence (varint),
 *   Frame Sequence (varint),
 * }
 */
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

	p = appendNumber(p, uint64(fm.SubscribeID))

	p = appendStringArray(p, fm.TrackPath)

	p = appendNumber(p, uint64(fm.TrackPriority))

	p = appendNumber(p, uint64(fm.GroupSequence))

	p = appendNumber(p, uint64(fm.FrameSequence))

	b := make([]byte, 0, len(p)+8)

	b = appendBytes(b, p)

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
	num, err := readNumber(mr)
	if err != nil {
		return err
	}
	fm.SubscribeID = SubscribeID(num)

	fm.TrackPath, err = readStringArray(mr)
	if err != nil {
		return err
	}

	// Get a Subscriber Priority
	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	fm.TrackPriority = TrackPriority(num)

	// Get a Group Sequence
	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	fm.GroupSequence = GroupSequence(num)

	// Get a Group Offset
	num, err = readNumber(mr)
	if err != nil {
		return err
	}
	fm.FrameSequence = FrameSequence(num)

	slog.Debug("decoded a FETCH message")

	return nil
}
