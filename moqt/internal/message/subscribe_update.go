package message

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * SUBSCRIBE_UPDATE Message {
 *   Track Priority (varint),
 *   Group Order (varint),
 *   Min Group Sequence (varint),
 *   Max Group Sequence (varint),
 *   Subscribe Update Parameters (Parameters),
 * }
 */
type SubscribeUpdateMessage struct {
	TrackPriority    TrackPriority
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence
}

func (su SubscribeUpdateMessage) Len() int {
	var l int

	l += VarintLen(uint64(su.TrackPriority))
	l += VarintLen(uint64(su.MinGroupSequence))
	l += VarintLen(uint64(su.MaxGroupSequence))

	return l
}

func (su SubscribeUpdateMessage) Encode(w io.Writer) error {
	msgLen := su.Len()
	p := pool.Get(msgLen)
	defer pool.Put(p)

	p = quicvarint.Append(p, uint64(msgLen))
	p = quicvarint.Append(p, uint64(su.TrackPriority))
	p = quicvarint.Append(p, uint64(su.MinGroupSequence))
	p = quicvarint.Append(p, uint64(su.MaxGroupSequence))

	_, err := w.Write(p)

	return err
}

func (sum *SubscribeUpdateMessage) Decode(src io.Reader) error {
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
	sum.TrackPriority = TrackPriority(num)
	b = b[n:]

	num, n, err = ReadVarint(b)
	if err != nil {
		return err
	}
	sum.MinGroupSequence = GroupSequence(num)
	b = b[n:]

	num, n, err = ReadVarint(b)
	if err != nil {
		return err
	}
	sum.MaxGroupSequence = GroupSequence(num)
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}
