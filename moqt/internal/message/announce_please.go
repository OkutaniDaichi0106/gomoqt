package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type AnnouncePleaseMessage struct {
	TrackPathPrefix []string
	Parameters      Parameters
}

func (aim AnnouncePleaseMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a ANNOUNCE_PLEASE message")

	// Serialize the message in the following format
	// ANNOUNCE_PLEASE Message Payload {
	//   Track Prefix ([]string),
	//   Announce Parameters (Parameters),
	// }

	// Serialize the payload
	p := make([]byte, 0, 1<<6) // TODO: Tune the size

	// Append the Track Namespace Prefix's length and parts
	p = appendStringArray(p, aim.TrackPathPrefix)

	// Append the Parameters
	p = appendParameters(p, aim.Parameters)

	// Get serialized message
	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))

	// Append the length of the payload and the payload itself
	b = appendBytes(b, p)

	// Write
	if _, err := w.Write(b); err != nil {
		return err
	}
	slog.Debug("encoded a ANNOUNCE_PLEASE message")

	return nil
}

func (aim *AnnouncePleaseMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a ANNOUNCE_PLEASE message")

	// Get a message reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get Track Namespace Prefix parts
	aim.TrackPathPrefix, err = readStringArray(mr)
	if err != nil {
		return err
	}

	// Get Parameters
	aim.Parameters, err = readParameters(mr)
	if err != nil {
		return err
	}

	slog.Debug("decoded a ANNOUNCE_PLEASE message")

	return nil
}
