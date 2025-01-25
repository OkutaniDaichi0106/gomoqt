package message

import (
	"bytes"
	"io"
	"log/slog"

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
	TrackPriority             TrackPriority
	GroupOrder                GroupOrder
	MinGroupSequence          GroupSequence
	MaxGroupSequence          GroupSequence
	SubscribeUpdateParameters Parameters
}

func (su SubscribeUpdateMessage) Encode(w io.Writer) (int, error) {
	slog.Debug("encoding a SUBSCRIBE_UPDATE message")

	// Serialize the payload
	p := make([]byte, 0, 1<<6)
	p = appendNumber(p, uint64(su.TrackPriority))
	p = appendNumber(p, uint64(su.GroupOrder))
	p = appendNumber(p, uint64(su.MinGroupSequence))
	p = appendNumber(p, uint64(su.MaxGroupSequence))
	p = appendParameters(p, su.SubscribeUpdateParameters)

	// Serialize the message
	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))
	b = appendBytes(b, p)

	// Write
	return w.Write(b)
}

func (sum *SubscribeUpdateMessage) Decode(r io.Reader) (int, error) {
	slog.Debug("decoding a SUBSCRIBE_UPDATE message")

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
	sum.TrackPriority = TrackPriority(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	sum.GroupOrder = GroupOrder(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	sum.MinGroupSequence = GroupSequence(num)

	num, _, err = readNumber(mr)
	if err != nil {
		return n, err
	}
	sum.MaxGroupSequence = GroupSequence(num)

	sum.SubscribeUpdateParameters, _, err = readParameters(mr)
	if err != nil {
		return n, err
	}

	slog.Debug("decoded a SUBSCRIBE_UPDATE message")
	return n, nil
}
