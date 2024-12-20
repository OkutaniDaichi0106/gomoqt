package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type GoAwayMessage struct {
	/*
	 * New session URI
	 * If this is 0 byte, this should be set to current session URI
	 */
	NewSessionURI string
}

func (ga GoAwayMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a GOAWAY message")

	/*
	 * Serialize the message in the following format
	 *
	 * GOAWAY Message Payload {
	 *   Message Length (varint),
	 *   New Session URI (string),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<6)

	// Append the supported versions
	p = quicvarint.Append(p, uint64(len(ga.NewSessionURI)))
	p = append(p, []byte(ga.NewSessionURI)...)

	// Get a serialized message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)
	if err != nil {
		slog.Error("failed to write a GOAWAY message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("encoded a GOAWAY message")

	return nil
}

func (ga *GoAwayMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a GOAWAY message")

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get length of the URI
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}

	// Get URI
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		slog.Error("failed to read an URI")
		return err
	}
	ga.NewSessionURI = string(buf)

	slog.Debug("decoded a GOAWAY message")

	return nil
}
