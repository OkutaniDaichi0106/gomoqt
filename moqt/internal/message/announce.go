package message

import (
	"bytes"
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

func (a AnnounceMessage) Encode(w io.Writer) (int, error) {
	slog.Debug("encoding a ANNOUNCE message")

	// Serialize the payload
	p := make([]byte, 0, 1<<6)

	p = appendNumber(p, uint64(a.AnnounceStatus))
	p = appendStringArray(p, a.TrackPathSuffix)
	p = appendParameters(p, a.Parameters)

	// Serialize the message
	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))
	b = appendBytes(b, p)

	// Write
	return w.Write(b)
}

func (am *AnnounceMessage) Decode(r io.Reader) (int, error) {
	// Read the payload
	buf, n, err := readBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	// Decode the payload
	mr := bytes.NewReader(buf)

	status, _, err := readNumber(mr)
	if err != nil {
		return n, err
	}
	am.AnnounceStatus = AnnounceStatus(status)

	am.TrackPathSuffix, _, err = readStringArray(mr)
	if err != nil {
		return n, err
	}

	am.Parameters, _, err = readParameters(mr)
	if err != nil {
		return n, err
	}

	return n, nil
}
