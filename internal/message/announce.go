package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	ended  byte = 0x0
	active byte = 0x1
	live   byte = 0x2
)

type AnnounceMessage struct {
	/*
	 * Announce Status
	 */
	AnnounceStatus byte

	/*
	 * Track Namespace
	 */
	TrackPath string

	/*
	 * Announce Parameters
	 * Parameters should include track authorization information
	 */
	Parameters Parameters
}

func (a AnnounceMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a ANNOUNCE message")

	/*
	 * Serialize the payload in the following format
	 *
	 * ANNOUNCE Message Payload {
	 *   Track Path (string),
	 *   Number of Parameters (),
	 *   Announce Parameters(..)
	 * }
	 */

	p := make([]byte, 0, 1<<6) // TODO: Tune the size

	// Append the Track Namespace
	p = quicvarint.Append(p, uint64(len(a.TrackPath)))
	p = append(p, []byte(a.TrackPath)...)

	// Append the Parameters
	p = appendParameters(p, a.Parameters)

	// Get serialized message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	_, err := w.Write(b)
	if err != nil {
		return err
	}

	slog.Debug("encoded a ANNOUNCE message")

	return nil
}

func (am *AnnounceMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a ANNOUNCE message")

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Track Path
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	am.TrackPath = string(buf)

	// Get Parameters
	am.Parameters, err = readParameters(mr)
	if err != nil {
		return err
	}

	slog.Debug("decoded a ANNOUNCE message")

	return nil
}
