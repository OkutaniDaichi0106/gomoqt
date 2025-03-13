package message

import (
	"bytes"
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
 *	ANNOUNCE_PLEASE Message {
 *	  Track Prefix ([]string),
 *	  Announce Parameters (Parameters),
 *	}
 */

type AnnouncePleaseMessage struct {
	TrackPattern string
	// AnnounceParameters Parameters
}

func (aim AnnouncePleaseMessage) Len() int {
	// Calculate the length of the payload
	l := 0
	l += stringLen(aim.TrackPattern)
	// l += parametersLen(aim.AnnounceParameters)

	return l
}

func (aim AnnouncePleaseMessage) Encode(w io.Writer) (int, error) {
	// Serialize the payload
	p := GetBytes()
	defer PutBytes(p)

	*p = AppendNumber(*p, uint64(aim.Len()))

	*p = AppendString(*p, aim.TrackPattern)
	// *p = AppendParameters(*p, aim.AnnounceParameters)

	n, err := w.Write(*p)
	if err != nil {
		slog.Error("failed to encode an ANNOUNCE_PLEASE message",
			"error", err,
		)
		return n, err
	}

	slog.Debug("encoded an ANNOUNCE_PLEASE message")

	return n, err
}

func (aim *AnnouncePleaseMessage) Decode(r io.Reader) (int, error) {

	// Read the payload
	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read payload for ANNOUNCE_PLEASE message",
			"error", err,
		)
		return n, err
	}

	// Decode the payload
	mr := bytes.NewReader(buf)

	aim.TrackPattern, _, err = ReadString(mr)
	if err != nil {
		slog.Error("failed to read TrackPrefix for ANNOUNCE_PLEASE message",
			"error", err,
		)
		return n, err
	}

	slog.Debug("decoded an ANNOUNCE_PLEASE message")

	return n, nil
}
