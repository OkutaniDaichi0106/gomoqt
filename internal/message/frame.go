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
	/*
	 * Serialize the message in following format
	 *
	 * Frame Message {
	 *   Payload Length (varint),
	 *   Payload ([]byte),
	 * }
	 */

	b := make([]byte, 0, len(fm.Payload)+8)

	// Append the payload length
	b = quicvarint.Append(b, uint64(len(fm.Payload)))

	// Append the payload
	b = append(b, fm.Payload...)

	_, err := w.Write(b)

	return err
}

func (fm *FrameMessage) Decode(r io.Reader) error {
	// Get a payload length
	num, err := quicvarint.Read(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to get a new message reader", slog.String("error", err.Error()))
		return err
	}

	// Get a payload
	buf := make([]byte, num)
	_, err = r.Read(buf)
	fm.Payload = buf

	return err
}
