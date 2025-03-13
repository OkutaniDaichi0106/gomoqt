package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type FrameSequence uint64

/*
 * Frame Message {
 *   Message Length (varint),
 *   Payload ([]byte),
 * }
 */

type FrameMessage struct {
	Payload []byte
}

func (fm FrameMessage) Len() int {
	return bytesLen(fm.Payload)
}

func (fm FrameMessage) Encode(w io.Writer) (int, error) {
	b := GetBytes()
	defer PutBytes(b)

	*b = AppendBytes(*b, fm.Payload)

	n, err := w.Write(*b)
	if err != nil {
		slog.Error("failed to write a FRAME message", "error", err)
		return n, err
	}

	slog.Debug("encoded a FRAME message")

	return n, nil
}

func (fm *FrameMessage) Decode(r io.Reader) (int, error) {
	var err error
	var n int

	fm.Payload, n, err = ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read payload for FRAME message", "error", err)
		return n, err
	}

	slog.Debug("decoded a FRAME message")

	return n, nil
}
