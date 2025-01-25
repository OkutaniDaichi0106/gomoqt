package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeID uint64
type TrackPriority byte
type GroupOrder byte

/*
 * SUBSCRIBE Message {
 *   Subscribe ID (varint),
 *   Track Path ([]string),
 *   Track Priority (varint),
 *   Group Order (varint),
 *   Min Group Sequence (varint),
 *   Max Group Sequence (varint),
 *   Subscribe Parameters (Parameters),
 * }
 */
type SubscribeMessage struct {
	SubscribeID         SubscribeID
	TrackPath           []string
	TrackPriority       TrackPriority
	GroupOrder          GroupOrder
	MinGroupSequence    GroupSequence
	MaxGroupSequence    GroupSequence
	SubscribeParameters Parameters
}

func (s SubscribeMessage) Encode(w io.Writer) (int, error) {
	// Serialize the payload
	p := make([]byte, 0, 1<<6)
	p = appendNumber(p, uint64(s.SubscribeID))
	p = appendStringArray(p, s.TrackPath)
	p = appendNumber(p, uint64(s.TrackPriority))
	p = appendNumber(p, uint64(s.GroupOrder))
	p = appendNumber(p, uint64(s.MinGroupSequence))
	p = appendNumber(p, uint64(s.MaxGroupSequence))
	p = appendParameters(p, s.SubscribeParameters)

	// Serialize the message
	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))
	b = appendBytes(b, p)

	return w.Write(b)
}

func (s *SubscribeMessage) Decode(r io.Reader) (int, error) {
	// Read the payload
	buf, n, err := readBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	// Decode the payload
	mr := bytes.NewReader(buf)

	num, _, err := readNumber(mr)
	if err != nil {
		return n, err
	}
	s.SubscribeID = SubscribeID(num)

	s.TrackPath, _, err = readStringArray(mr)
	if err != nil {
		return n, err
	}

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	s.TrackPriority = TrackPriority(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	s.GroupOrder = GroupOrder(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	s.MinGroupSequence = GroupSequence(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	s.MaxGroupSequence = GroupSequence(num)

	s.SubscribeParameters, _, err = readParameters(mr)
	if err != nil {
		return n, err
	}

	return n, nil
}
