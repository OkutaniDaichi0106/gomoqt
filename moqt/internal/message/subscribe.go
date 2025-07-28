package message

import (
	"bytes"
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeID = protocol.SubscribeID
type TrackPriority = protocol.TrackPriority
type GroupOrder = protocol.GroupOrder

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
	var l int

	l += VarintLen(uint64(s.SubscribeID))
	l += StringLen(s.BroadcastPath)
	l += StringLen(s.TrackName)
	l += VarintLen(uint64(s.TrackPriority))
	l += VarintLen(uint64(s.MinGroupSequence))
	l += VarintLen(uint64(s.MaxGroupSequence))

	return l
}

func (s SubscribeMessage) Encode(w io.Writer) error {
	msgLen := s.Len()
	b := pool.Get(msgLen)
	defer pool.Put(b)

	b = quicvarint.Append(b, uint64(msgLen))
	b = quicvarint.Append(b, uint64(s.SubscribeID))
	b = quicvarint.Append(b, uint64(len(s.BroadcastPath)))
	b = append(b, s.BroadcastPath...)
	b = quicvarint.Append(b, uint64(len(s.TrackName)))
	b = append(b, s.TrackName...)
	b = quicvarint.Append(b, uint64(s.TrackPriority))
	b = quicvarint.Append(b, uint64(s.MinGroupSequence))
	b = quicvarint.Append(b, uint64(s.MaxGroupSequence))

	_, err := w.Write(b)
	return err
}

func (s *SubscribeMessage) Decode(src io.Reader) error {
	num, err := ReadVarint(src)
	if err != nil {
		return err
	}

	b := pool.Get(int(num))[:num]
	defer pool.Put(b)

	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	r := bytes.NewReader(b)

	num, err = ReadVarint(r)
	if err != nil {
		return err
	}
	s.SubscribeID = SubscribeID(num)

	str, err := ReadString(r)
	if err != nil {
		return err
	}
	s.BroadcastPath = str

	str, err = ReadString(r)
	if err != nil {
		return err
	}
	s.TrackName = str

	num, err = ReadVarint(r)
	if err != nil {
		return err
	}
	s.TrackPriority = TrackPriority(num)

	num, err = ReadVarint(r)
	if err != nil {
		return err
	}
	s.MinGroupSequence = GroupSequence(num)

	num, err = ReadVarint(r)
	if err != nil {
		return err
	}
	s.MaxGroupSequence = GroupSequence(num)

	return nil
}
