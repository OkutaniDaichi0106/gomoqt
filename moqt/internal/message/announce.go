package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	ENDED  AnnounceStatus = 0x0
	ACTIVE AnnounceStatus = 0x1
	LIVE   AnnounceStatus = 0x2
)

type AnnounceStatus byte

type AnnounceMessage struct {
	AnnounceStatus  AnnounceStatus
	TrackPathSuffix []string
	Parameters      Parameters
}

func (a AnnounceMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a ANNOUNCE message")

	// Serialize the payload in the following format
	// ANNOUNCE Message Payload {
	//   Track Path (string),
	//   Number of Parameters (),
	//   Announce Parameters(..)
	// }

	p := make([]byte, 0, 1<<6) // TODO: Tune the size

	// Append the Announce Status
	p = appendNumber(p, uint64(a.AnnounceStatus))

	// Append the Track Path Suffix's length and parts
	p = appendStringArray(p, a.TrackPathSuffix)

	// Append the Parameters
	p = appendParameters(p, a.Parameters)

	// Get serialized message
	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))

	// Append the length of the payload and the payload itself
	b = appendBytes(b, p)

	// Write
	_, err := w.Write(b)
	if err != nil {
		return err
	}

	slog.Debug("encoded a ANNOUNCE message")

	return nil
}

func (am *AnnounceMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a ANNOUNCE message")

	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get an Announce Status
	status, err := readNumber(mr)
	if err != nil {
		return err
	}
	am.AnnounceStatus = AnnounceStatus(status)

	// Get Track Path Suffix parts
	am.TrackPathSuffix, err = readStringArray(mr)
	if err != nil {
		return err
	}

	// Get Parameters
	am.Parameters, err = readParameters(mr)
	if err != nil {
		return err
	}

	slog.Debug("decoded a ANNOUNCE message")

	return nil
}
