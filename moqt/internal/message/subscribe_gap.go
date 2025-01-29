package message

import (
	"bytes"
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type GroupErrorCode uint64

/*
 * SUBSCRIBE_GAP Message {
 *   Min Gap Sequence (varint),
 *   Max Gap Sequence (varint),
 *   Group Error Code (varint),
 * }
 */
type SubscribeGapMessage struct {
	MinGapSequence GroupSequence
	MaxGapSequence GroupSequence
	GroupErrorCode GroupErrorCode
}

func (sgm SubscribeGapMessage) Len() int {
	l := 0
	l += numberLen(uint64(sgm.MinGapSequence))
	l += numberLen(uint64(sgm.MaxGapSequence))
	l += numberLen(uint64(sgm.GroupErrorCode))
	return l
}

func (sgm SubscribeGapMessage) Encode(w io.Writer) (int, error) {
	slog.Debug("encoding a SUBSCRIBE_GAP message")

	p := GetBytes()
	defer PutBytes(p)

	*p = AppendNumber(*p, uint64(sgm.Len()))
	*p = AppendNumber(*p, uint64(sgm.MinGapSequence))
	*p = AppendNumber(*p, uint64(sgm.MaxGapSequence))
	*p = AppendNumber(*p, uint64(sgm.GroupErrorCode))

	n, err := w.Write(*p)
	if err != nil {
		slog.Error("failed to write a SUBSCRIBE_GAP message", slog.String("error", err.Error()))
		return n, err
	}

	slog.Debug("encoded a SUBSCRIBE_GAP message", slog.Int("bytes_written", n))

	return n, nil
}

func (sgm *SubscribeGapMessage) Decode(r io.Reader) (int, error) {
	slog.Debug("decoding a SUBSCRIBE_GAP message")

	// Read the payload
	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	// Decode the payload
	mr := bytes.NewReader(buf)

	num, _, err := ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read MinGapSequence for SUBSCRIBE_GAP message", slog.String("error", err.Error()))
		return n, err
	}
	sgm.MinGapSequence = GroupSequence(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read MaxGapSequence for SUBSCRIBE_GAP message", slog.String("error", err.Error()))
		return n, err
	}
	sgm.MaxGapSequence = GroupSequence(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read GroupErrorCode for SUBSCRIBE_GAP message", slog.String("error", err.Error()))
		return n, err
	}
	sgm.GroupErrorCode = GroupErrorCode(num)

	slog.Debug("decoded a SUBSCRIBE_GAP message")

	return n, nil
}
