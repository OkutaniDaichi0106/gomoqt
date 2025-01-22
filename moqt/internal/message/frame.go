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

func (fm FrameMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a FRAME message")

	/*
	 * Serialize the message in following format
	 *
	 * Frame Message {
	 *   Message Length (varint),
	 *   Payload ([]byte),
	 * }
	 */

	b := make([]byte, 0, len(fm.Payload)+8)

	// Append the payload length
	b = quicvarint.Append(b, uint64(len(fm.Payload)))

	// Append the payload
	b = append(b, fm.Payload...)

	_, err := w.Write(b)
	if err != nil {
		slog.Error("failed to write a FRAME message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("encoded a FRAME message")

	return nil
}

func (fm *FrameMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a FRAME message")

	mr, err := newReader(r)
	if err != nil {
		return err
	}

	fm.Payload, err = io.ReadAll(mr)
	if err != nil {
		return err
	}

	slog.Debug("decoded a FRAME message")

	return nil
}
