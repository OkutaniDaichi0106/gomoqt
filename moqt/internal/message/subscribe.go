package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeID uint64
type TrackPriority byte
type GroupOrder byte

const (
	GroupOrderDefault    GroupOrder = 0x00
	GroupOrderAscending  GroupOrder = 0x01
	GroupOrderDescending GroupOrder = 0x02
)

/*
* SUBSCRIBE Message {
*   Subscribe ID (varint),
*   Broadcast Path (string),
*   Track Name (string),
*   Track Priority (varint),
*   Min Group Sequence (varint),
*   Max Group Sequence (varint),
* }
 */
type SubscribeMessage struct {
	SubscribeID      SubscribeID
	BroadcastPath    string
	TrackName        string
	TrackPriority    TrackPriority
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence
}

func (s SubscribeMessage) Len() int {
	l := 0
	l += numberLen(uint64(s.SubscribeID))
	l += stringLen(s.BroadcastPath)
	l += stringLen(s.TrackName)
	l += numberLen(uint64(s.TrackPriority))
	l += numberLen(uint64(s.MinGroupSequence))
	l += numberLen(uint64(s.MaxGroupSequence))
	return l
}

func (s SubscribeMessage) Encode(w io.Writer) (int, error) {

	p := GetBytes()
	defer PutBytes(p)

	p = AppendNumber(p, uint64(s.Len()))

	p = AppendNumber(p, uint64(s.SubscribeID))
	p = AppendString(p, s.BroadcastPath)
	p = AppendString(p, s.TrackName)
	p = AppendNumber(p, uint64(s.TrackPriority))
	p = AppendNumber(p, uint64(s.MinGroupSequence))
	p = AppendNumber(p, uint64(s.MaxGroupSequence))

	n, err := w.Write(p)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (s *SubscribeMessage) Decode(r io.Reader) (int, error) {

	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := ReadNumber(mr)
	if err != nil {

		return n, err
	}
	s.SubscribeID = SubscribeID(num)

	s.BroadcastPath, _, err = ReadString(mr)
	if err != nil {
		return n, err
	}

	s.TrackName, _, err = ReadString(mr)
	if err != nil {
		return n, err
	}

	num, _, err = ReadNumber(mr)
	if err != nil {
		return n, err
	}
	s.TrackPriority = TrackPriority(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		return n, err
	}
	s.MinGroupSequence = GroupSequence(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		return n, err
	}
	s.MaxGroupSequence = GroupSequence(num)

	return n, nil
}
