package message

import (
	"io"
	"log"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type InfoRequestMessage struct {
	/*
	 * Track name
	 */
	TrackPath []string
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

	// Append the Track Path Length
	p = quicvarint.Append(p, uint64(len(irm.TrackPath)))

	// Append the Track Path
	for _, part := range irm.TrackPath {
		// Append the Track Namespace Prefix Part
		p = quicvarint.Append(p, uint64(len(part)))
		p = append(p, []byte(part)...)
	}

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

	// Get a Track Path Part Count
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}

	// Get a Track Path
	irm.TrackPath = make([]string, num)

	// Get Track Path Parts
	for i := 0; i < int(num); i++ {
		num, err = quicvarint.Read(mr)
		if err != nil {
			return err
		}

		// Get a Track Path Part
		buf := make([]byte, num)
		_, err = r.Read(buf)
		if err != nil {
			return err
		}

		irm.TrackPath[i] = string(buf)
	}

	slog.Debug("decoded a INFO_REQUEST message")

	return nil
}
