package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
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

func (fm FetchMessage) Encode(w io.Writer) (int, error) {
	p := make([]byte, 0, 1<<8)
	p = appendNumber(p, uint64(fm.SubscribeID))
	p = appendStringArray(p, fm.TrackPath)
	p = appendNumber(p, uint64(fm.TrackPriority))
	p = appendNumber(p, uint64(fm.GroupSequence))
	p = appendNumber(p, uint64(fm.FrameSequence))

	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))
	b = appendBytes(b, p)

	return w.Write(b)
}

func (fm *FetchMessage) Decode(r io.Reader) (int, error) {
	buf, n, err := readBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := readNumber(mr)
	if err != nil {
		return n, err
	}
	fm.SubscribeID = SubscribeID(num)

	fm.TrackPath, _, err = readStringArray(mr)
	if err != nil {
		return n, err
	}

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	fm.TrackPriority = TrackPriority(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	fm.GroupSequence = GroupSequence(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	fm.FrameSequence = FrameSequence(num)

	return n, nil
}
