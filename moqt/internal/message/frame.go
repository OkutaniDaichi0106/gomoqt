package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type FrameSequence uint64

type FrameMessage struct {
	Payload []byte
}

func (fm FrameMessage) Encode(w io.Writer) (int, error) {
	slog.Debug("encoding a FRAME message")

	/*
	 * Serialize the message in following format
	 *
	 * Frame Message {
	 *   Message Length (varint),
	 *   Payload ([]byte),
	 * }
	 */

	b := make([]byte, 0, len(fm.Payload)+quicvarint.Len(uint64(len(fm.Payload))))

	// Append the payload length
	b = appendBytes(b, fm.Payload)

	// Write
	n, err := w.Write(b)
	if err != nil {
		slog.Error("failed to write a FRAME message", slog.String("error", err.Error()))
		return n, err
	}

	slog.Debug("encoded a FRAME message")

	return n, nil
}

func (fm *FrameMessage) Decode(r io.Reader) (int, error) {
	slog.Debug("decoding a FRAME message")

	var err error
	var n int

	fm.Payload, n, err = readBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	slog.Debug("decoded a FRAME message")

	return n, nil
}
