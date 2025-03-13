package message

import (
	"bytes"
	"io"
	"log/slog"

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
*   Track Path ([]string),
*   Track Priority (varint),
*   Group Order (varint),
*   Min Group Sequence (varint),
*   Max Group Sequence (varint),
*   // Subscribe Parameters (Parameters),
* }
 */
type SubscribeMessage struct {
	SubscribeID      SubscribeID
	TrackPath        string
	TrackPriority    TrackPriority
	GroupOrder       GroupOrder
	MinGroupSequence GroupSequence
	MaxGroupSequence GroupSequence
	// SubscribeParameters Parameters
}

func (s SubscribeMessage) Len() int {
	l := 0
	l += numberLen(uint64(s.SubscribeID))
	l += stringLen(s.TrackPath)
	l += numberLen(uint64(s.TrackPriority))
	l += numberLen(uint64(s.GroupOrder))
	l += numberLen(uint64(s.MinGroupSequence))
	l += numberLen(uint64(s.MaxGroupSequence))
	// l += parametersLen(s.SubscribeParameters)
	return l
}

func (s SubscribeMessage) Encode(w io.Writer) (int, error) {

	p := GetBytes()
	defer PutBytes(p)

	*p = AppendNumber(*p, uint64(s.Len()))

	*p = AppendNumber(*p, uint64(s.SubscribeID))
	*p = AppendString(*p, s.TrackPath)
	*p = AppendNumber(*p, uint64(s.TrackPriority))
	*p = AppendNumber(*p, uint64(s.GroupOrder))
	*p = AppendNumber(*p, uint64(s.MinGroupSequence))
	*p = AppendNumber(*p, uint64(s.MaxGroupSequence))

	n, err := w.Write(*p)
	if err != nil {
		slog.Error("failed to write a SUBSCRIBE message", "error", err)
		return n, err
	}

	slog.Debug("encoded a SUBSCRIBE message", slog.Int("bytes_written", n))

	return n, nil
}

func (s *SubscribeMessage) Decode(r io.Reader) (int, error) {

	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read payload for SUBSCRIBE message",
			"error", err,
		)
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read subscribe ID for SUBSCRIBE message",
			"error", err,
		)
		return n, err
	}
	s.SubscribeID = SubscribeID(num)

	s.TrackPath, _, err = ReadString(mr)
	if err != nil {
		slog.Error("failed to read track path for SUBSCRIBE message",
			"error", err,
		)
		return n, err
	}

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read track priority for SUBSCRIBE message",
			"error", err,
		)
		return n, err
	}
	s.TrackPriority = TrackPriority(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read group order for SUBSCRIBE message",
			"error", err,
		)
		return n, err
	}
	s.GroupOrder = GroupOrder(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read min group sequence for SUBSCRIBE message",
			"error", err,
		)
		return n, err
	}
	s.MinGroupSequence = GroupSequence(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read max group sequence for SUBSCRIBE message",
			"error", err,
		)
		return n, err
	}
	s.MaxGroupSequence = GroupSequence(num)

	slog.Debug("decoded a SUBSCRIBE message")

	return n, nil
}
