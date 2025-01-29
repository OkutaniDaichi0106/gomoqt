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

func (fm FetchMessage) Len() int {
	l := 0
	l += numberLen(uint64(fm.SubscribeID))
	l += stringArrayLen(fm.TrackPath)
	l += numberLen(uint64(fm.TrackPriority))
	l += numberLen(uint64(fm.GroupSequence))
	l += numberLen(uint64(fm.FrameSequence))
	return l
}

func (fm FetchMessage) Encode(w io.Writer) (int, error) {
	p := GetBytes()
	defer PutBytes(p)

	*p = AppendNumber(*p, uint64(fm.Len()))
	*p = AppendNumber(*p, uint64(fm.SubscribeID))
	*p = AppendStringArray(*p, fm.TrackPath)
	*p = AppendNumber(*p, uint64(fm.TrackPriority))
	*p = AppendNumber(*p, uint64(fm.GroupSequence))
	*p = AppendNumber(*p, uint64(fm.FrameSequence))

	return w.Write(*p)
}

func (fm *FetchMessage) Decode(r io.Reader) (int, error) {
	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := ReadNumber(mr)
	if err != nil {
		return n, err
	}
	fm.SubscribeID = SubscribeID(num)

	fm.TrackPath, _, err = ReadStringArray(mr)
	if err != nil {
		return n, err
	}

	num, _, err = ReadNumber(mr)
	if err != nil {
		return n, err
	}
	fm.TrackPriority = TrackPriority(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		return n, err
	}
	fm.GroupSequence = GroupSequence(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		return n, err
	}
	fm.FrameSequence = FrameSequence(num)

	return n, nil
}
