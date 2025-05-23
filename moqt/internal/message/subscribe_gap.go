package message

import (
	"bytes"
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeGapMessage struct {
	StartGroupSequence GroupSequence
	GapCount           uint64
	GroupErrorCode     protocol.GroupErrorCode
}

func (s SubscribeGapMessage) Len() int {
	l := 0
	l += numberLen(uint64(s.StartGroupSequence))
	l += numberLen(uint64(s.GapCount))
	l += numberLen(uint64(s.GroupErrorCode))
	return l
}

func (s SubscribeGapMessage) Encode(w io.Writer) (int, error) {

	p := GetBytes()
	defer PutBytes(p)

	p = AppendNumber(p, uint64(s.Len()))

	p = AppendNumber(p, uint64(s.StartGroupSequence))
	p = AppendNumber(p, uint64(s.GapCount))
	p = AppendNumber(p, uint64(s.GroupErrorCode))

	n, err := w.Write(p)
	if err != nil {
		slog.Error("failed to write a SUBSCRIBE_GAP message", "error", err)
		return n, err
	}

	slog.Debug("encoded a SUBSCRIBE_GAP message", slog.Int("bytes_written", n))

	return n, nil
}

func (s *SubscribeGapMessage) Decode(r io.Reader) (int, error) {

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
	s.StartGroupSequence = GroupSequence(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read track priority for SUBSCRIBE message",
			"error", err,
		)
		return n, err
	}
	s.GapCount = num

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read group order for SUBSCRIBE message",
			"error", err,
		)
		return n, err
	}
	s.GroupErrorCode = protocol.GroupErrorCode(num)

	slog.Debug("decoded a SUBSCRIBE_GAP message")

	return n, nil
}
