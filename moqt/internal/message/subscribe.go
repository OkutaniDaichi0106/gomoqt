package message

import (
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

func (s SubscribeMessage) Encode(w io.Writer) error {

	p := getBytes()
	defer putBytes(p)

	p = AppendNumber(p, uint64(s.SubscribeID))
	p = AppendString(p, s.BroadcastPath)
	p = AppendString(p, s.TrackName)
	p = AppendNumber(p, uint64(s.TrackPriority))
	p = AppendNumber(p, uint64(s.MinGroupSequence))
	p = AppendNumber(p, uint64(s.MaxGroupSequence))

	_, err := w.Write(p)
	return err
}

func (s *SubscribeMessage) Decode(r io.Reader) error {
	num, _, err := ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}
	s.SubscribeID = SubscribeID(num)

	s.BroadcastPath, _, err = ReadString(quicvarint.NewReader(r))
	if err != nil {
		return err
	}

	s.TrackName, _, err = ReadString(quicvarint.NewReader(r))
	if err != nil {
		return err
	}

	num, _, err = ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}
	s.TrackPriority = TrackPriority(num)

	num, _, err = ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}
	s.MinGroupSequence = GroupSequence(num)

	num, _, err = ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}
	s.MaxGroupSequence = GroupSequence(num)

	return nil
}
