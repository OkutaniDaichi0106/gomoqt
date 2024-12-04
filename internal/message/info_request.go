package message

import (
	"io"
	"log"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type InfoRequestMessage struct {
	/*
	 * Track namespace
	 */
	TrackNamespace string

	/*
	 * Track name
	 */
	TrackName string
}

func (irm InfoRequestMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a INFO_REQUEST message")

	/*
	 * Serialize the payload in the following format
	 *
	 * TRACK_STATUS_REQUEST Message Payload {
	 *   Track Namespace (tuple),
	 *   Track Name ([]byte),
	 * }
	 */

	p := make([]byte, 0, 1<<8)

	// Append the Track Namespace
	p = quicvarint.Append(p, uint64(len(irm.TrackNamespace)))
	p = append(p, []byte(irm.TrackNamespace)...)

	// Append the Track Name
	p = quicvarint.Append(p, uint64(len(irm.TrackName)))
	p = append(p, []byte(irm.TrackName)...)

	log.Print("INFO_REQUEST payload", p)

	// Serialize the whole message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)
	if err != nil {
		slog.Error("failed to write a INFO_REQUEST message")
		return err
	}

	slog.Debug("encoded a INFO_REQUEST message")

	return nil
}

func (irm *InfoRequestMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a INFO_REQUEST message")

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Track Namespace
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	irm.TrackNamespace = string(buf)

	// Get a Track Name
	num, err = quicvarint.Read(mr)
	if err != nil {
		return err
	}
	buf = make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	irm.TrackName = string(buf)

	slog.Debug("decoded a INFO_REQUEST message")

	return nil
}
