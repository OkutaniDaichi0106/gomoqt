package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type SessionUpdateMessage struct {
	/*
	 * Versions selected by the server
	 */
	Bitrate uint64
}

func (sum SessionUpdateMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a SESSION_UPDATE message")

	/*
	 * Serialize the message in the following format
	 *
	 * SESSION_UPDATE Message Payload {
	 *   Bitrate (varint),
	 * }
	 */
	p := make([]byte, 0, 1<<3)

	// Append the Bitrate
	p = quicvarint.Append(p, sum.Bitrate)

	// Get a serialzed message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)

	return err
}

func (sum *SessionUpdateMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a SESSION_UPDATE message")

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a bitrate
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	sum.Bitrate = num

	return nil
}
