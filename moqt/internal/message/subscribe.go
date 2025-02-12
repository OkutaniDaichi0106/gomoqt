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

func (s SubscribeMessage) Len() int {
	l := 0
	l += numberLen(uint64(s.SubscribeID))
	l += stringArrayLen(s.TrackPath)
	l += numberLen(uint64(s.TrackPriority))
	l += numberLen(uint64(s.GroupOrder))
	l += numberLen(uint64(s.MinGroupSequence))
	l += numberLen(uint64(s.MaxGroupSequence))
	l += parametersLen(s.SubscribeParameters)
	return l
}

func (s SubscribeMessage) Encode(w io.Writer) (int, error) {
	slog.Debug("encoding a SUBSCRIBE message")

	p := GetBytes()
	defer PutBytes(p)

	*p = AppendNumber(*p, uint64(s.Len()))

	*p = AppendNumber(*p, uint64(s.SubscribeID))
	*p = AppendStringArray(*p, s.TrackPath)
	*p = AppendNumber(*p, uint64(s.TrackPriority))
	*p = AppendNumber(*p, uint64(s.GroupOrder))
	*p = AppendNumber(*p, uint64(s.MinGroupSequence))
	*p = AppendNumber(*p, uint64(s.MaxGroupSequence))

	*p = AppendParameters(*p, s.SubscribeParameters)

	n, err := w.Write(*p)
	if err != nil {
		slog.Error("failed to write a SUBSCRIBE message", slog.String("error", err.Error()))
		return n, err
	}

	slog.Debug("encoded a SUBSCRIBE message", slog.Int("bytes_written", n))

	return n, nil
}

func (s *SubscribeMessage) Decode(r io.Reader) (int, error) {
	slog.Debug("decoding a SUBSCRIBE message")

	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read bytes for SUBSCRIBE message", slog.String("error", err.Error()), slog.Int("bytes_read", n))
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read SubscribeID for SUBSCRIBE message", slog.String("error", err.Error()))
		return n, err
	}
	s.SubscribeID = SubscribeID(num)

	s.TrackPath, _, err = ReadStringArray(mr)
	if err != nil {
		slog.Error("failed to read TrackPath for SUBSCRIBE message", slog.String("error", err.Error()))
		return n, err
	}

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read TrackPriority for SUBSCRIBE message", slog.String("error", err.Error()))
		return n, err
	}
	s.TrackPriority = TrackPriority(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read GroupOrder for SUBSCRIBE message", slog.String("error", err.Error()))
		return n, err
	}
	s.GroupOrder = GroupOrder(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read MinGroupSequence for SUBSCRIBE message", slog.String("error", err.Error()))
		return n, err
	}
	s.MinGroupSequence = GroupSequence(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read MaxGroupSequence for SUBSCRIBE message", slog.String("error", err.Error()))
		return n, err
	}
	s.MaxGroupSequence = GroupSequence(num)

	s.SubscribeParameters, _, err = ReadParameters(mr)
	if err != nil {
		slog.Error("failed to read SubscribeParameters for SUBSCRIBE message", slog.String("error", err.Error()))
		return n, err
	}

	slog.Debug("decoded a SUBSCRIBE message")

	return n, nil
}
