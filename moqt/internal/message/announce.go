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
	AnnounceStatus AnnounceStatus
	TrackSuffix    string
	// AnnounceParameters Parameters
}

func (a AnnounceMessage) Len() int {
	l := 0
	l += numberLen(uint64(a.AnnounceStatus))
	l += stringLen(a.TrackSuffix)
	// l += parametersLen(a.AnnounceParameters)
	return l
}

func (a AnnounceMessage) Encode(w io.Writer) (int, error) {

	p := GetBytes()
	defer PutBytes(p)

	p = AppendNumber(p, uint64(a.Len()))
	p = AppendNumber(p, uint64(a.AnnounceStatus))
	p = AppendString(p, a.TrackSuffix)

	n, err := w.Write(p)
	if err != nil {
		slog.Error("failed to write an ANNOUNCE message",
			"error", err,
		)
		return n, err
	}

	slog.Debug("encoded an ANNOUNCE message")

	return n, nil
}

func (am *AnnounceMessage) Decode(r io.Reader) (int, error) {

	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read payload for ANNOUNCE message",
			"error", err,
		)
		return n, err
	}

	mr := bytes.NewReader(buf)

	status, _, err := ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read announce status for ANNOUNCE message",
			"error", err,
		)
		return n, err
	}
	am.AnnounceStatus = AnnounceStatus(status)

	am.TrackSuffix, _, err = ReadString(mr)
	if err != nil {
		slog.Error("failed to read track suffix for ANNOUNCE message",
			"error", err,
		)
		return n, err
	}

	slog.Debug("decoded an ANNOUNCE message")

	return n, nil
}
