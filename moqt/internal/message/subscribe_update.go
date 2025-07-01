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
	l := 0
	l += numberLen(uint64(su.TrackPriority))
	l += numberLen(uint64(su.MinGroupSequence))
	l += numberLen(uint64(su.MaxGroupSequence))

	return l
}

func (su SubscribeUpdateMessage) Encode(w io.Writer) error {

	p := getBytes()
	defer putBytes(p)

	p = AppendNumber(p, uint64(su.TrackPriority))
	p = AppendNumber(p, uint64(su.MinGroupSequence))
	p = AppendNumber(p, uint64(su.MaxGroupSequence))

	_, err := w.Write(p)
	return err
}

func (sum *SubscribeUpdateMessage) Decode(r io.Reader) error {
	num, _, err := ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}
	sum.TrackPriority = TrackPriority(num)

	num, _, err = ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}
	sum.MinGroupSequence = GroupSequence(num)

	num, _, err = ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}
	sum.MaxGroupSequence = GroupSequence(num)

	return nil
}
