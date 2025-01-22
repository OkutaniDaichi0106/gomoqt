package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceInterestMessage struct {
	TrackPathPrefix []string
	Parameters      Parameters
}

func (aim AnnounceInterestMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a ANNOUNCE_INTEREST message")

	/*
	 * Serialize the message in the following formatt
	 *
	 * ANNOUNCE_INTEREST Message Payload {
	 *   Track Namespace prefix (tuple),
	 *   Subscribe Parameters (Parameters),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<6) // TODO: Tune the size

	// Append the Track Namespace Prefix's length
	p = quicvarint.Append(p, uint64(len(aim.TrackPathPrefix)))

	for _, part := range aim.TrackPathPrefix {
		// Append the Track Namespace Prefix Part
		p = quicvarint.Append(p, uint64(len(part)))
		p = append(p, []byte(part)...)
	}

	// Append the Parameters
	p = appendParameters(p, aim.Parameters)

	/*
	 * Get serialized message
	 */
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)
	if err != nil {
		return err
	}
	slog.Debug("encoded a ANNOUNCE_INTEREST message")

	return nil
}

func (aim *AnnounceInterestMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a ANNOUNCE_INTEREST message")

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Track Namespace Prefix
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}

	// Get a Track Namespace Prefix Parts
	aim.TrackPathPrefix = make([]string, 0, num)

	// Get a Track Namespace Prefix Parts
	for i := 0; i < int(num); i++ {
		// Get a Track Namespace Prefix Part
		num, err = quicvarint.Read(mr)
		if err != nil {
			return err
		}

		buf := make([]byte, num)
		_, err = r.Read(buf)
		if err != nil {
			return err
		}
		aim.TrackPathPrefix = append(aim.TrackPathPrefix, string(buf))
	}

	// Get Parameters
	aim.Parameters, err = readParameters(mr)
	if err != nil {
		return err
	}

	slog.Debug("decoded a ANNOUNCE_INTEREST message")

	return nil
}
