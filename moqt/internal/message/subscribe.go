package message

import (
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeID = protocol.SubscribeID
type TrackPriority = protocol.TrackPriority

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
	b := pool.Get(msgLen + VarintLen(uint64(msgLen)))
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
	num, err := ReadMessageLength(src)
	if err != nil {
		return err
	}

	b := pool.Get(int(num))[:num]
	defer pool.Put(b)

	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	num, n, err := ReadVarint(b)
	if err != nil {
		return err
	}
	s.SubscribeID = SubscribeID(num)
	b = b[n:]

	str, n, err := ReadString(b)
	if err != nil {
		return err
	}
	s.BroadcastPath = str
	b = b[n:]

	str, n, err = ReadString(b)
	if err != nil {
		return err
	}
	s.TrackName = str
	b = b[n:]

	num, n, err = ReadVarint(b)
	if err != nil {
		return err
	}
	s.TrackPriority = TrackPriority(num)
	b = b[n:]

	num, n, err = ReadVarint(b)
	if err != nil {
		return err
	}
	s.MinGroupSequence = GroupSequence(num)
	b = b[n:]

	num, n, err = ReadVarint(b)
	if err != nil {
		return err
	}
	s.MaxGroupSequence = GroupSequence(num)
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}
