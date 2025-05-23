package message

import (
	"bytes"
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * SUBSCRIBE_OK Message {
 *   Group Order (varint),
 * }
 */
type SubscribeOkMessage struct {
	GroupOrder GroupOrder
}

func (som SubscribeOkMessage) Len() int {
	return numberLen(uint64(som.GroupOrder))
}

func (som SubscribeOkMessage) Encode(w io.Writer) (int, error) {
	p := GetBytes()
	defer PutBytes(p)

	p = AppendNumber(p, uint64(som.Len()))
	p = AppendNumber(p, uint64(som.GroupOrder))

	n, err := w.Write(p)
	if err != nil {
		slog.Error("failed to write a SUBSCRIBE_OK message", "error", err)
		return n, err
	}

	slog.Debug("encoded a SUBSCRIBE_OK message", slog.Int("bytes_written", n))

	return n, nil
}

func (som *SubscribeOkMessage) Decode(r io.Reader) (int, error) {
	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read payload for SUBSCRIBE_OK message", "error", err)
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read GroupOrder for SUBSCRIBE_OK message", "error", err)
		return n, err
	}
	som.GroupOrder = GroupOrder(num)

	slog.Debug("decoded a SUBSCRIBE_OK message")

	return n, nil
}
