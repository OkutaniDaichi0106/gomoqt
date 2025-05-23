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

func (su SubscribeUpdateMessage) Encode(w io.Writer) (int, error) {

	p := GetBytes()
	defer PutBytes(p)

	p = AppendNumber(p, uint64(su.Len()))
	p = AppendNumber(p, uint64(su.TrackPriority))
	p = AppendNumber(p, uint64(su.MinGroupSequence))
	p = AppendNumber(p, uint64(su.MaxGroupSequence))

	n, err := w.Write(p)
	if err != nil {
		slog.Error("failed to write a SUBSCRIBE_UPDATE message",
			"error", err,
		)
		return n, err
	}

	slog.Debug("encoded a SUBSCRIBE_UPDATE message", slog.Int("bytes_written", n))

	return n, nil
}

func (sum *SubscribeUpdateMessage) Decode(r io.Reader) (int, error) {

	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read payload for SUBSCRIBE_UPDATE message", "error", err)
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read TrackPriority for SUBSCRIBE_UPDATE message", "error", err)
		return n, err
	}
	sum.TrackPriority = TrackPriority(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read MinGroupSequence for SUBSCRIBE_UPDATE message", "error", err)
		return n, err
	}
	sum.MinGroupSequence = GroupSequence(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read MaxGroupSequence for SUBSCRIBE_UPDATE message", "error", err)
		return n, err
	}
	sum.MaxGroupSequence = GroupSequence(num)

	slog.Debug("decoded a SUBSCRIBE_UPDATE message")

	return n, nil
}
