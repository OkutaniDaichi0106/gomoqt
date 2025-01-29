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
	TrackPathPrefix []string
	Parameters      Parameters
}

func (aim AnnouncePleaseMessage) Len() int {
	// Calculate the length of the payload
	l := 0
	l += stringArrayLen(aim.TrackPathPrefix)
	l += parametersLen(aim.Parameters)

	return l
}

func (aim AnnouncePleaseMessage) Encode(w io.Writer) (int, error) {
	slog.Debug("encoding an ANNOUNCE_PLEASE message")

	// Serialize the payload
	p := GetBytes()
	defer PutBytes(p)

	*p = AppendNumber(*p, uint64(aim.Len()))

	*p = AppendStringArray(*p, aim.TrackPathPrefix)
	*p = AppendParameters(*p, aim.Parameters)

	n, err := w.Write(*p)
	if err != nil {
		slog.Error("failed to write an ANNOUNCE_PLEASE message", slog.String("error", err.Error()))
		return n, err
	}

	slog.Debug("encoded an ANNOUNCE_PLEASE message")

	return n, err
}

func (aim *AnnouncePleaseMessage) Decode(r io.Reader) (int, error) {
	slog.Debug("decoding an ANNOUNCE_PLEASE message")

	// Read the payload
	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read payload for ANNOUNCE_PLEASE message", slog.String("error", err.Error()))
		return n, err
	}

	// Decode the payload
	mr := bytes.NewReader(buf)

	aim.TrackPathPrefix, _, err = ReadStringArray(mr)
	if err != nil {
		slog.Error("failed to read TrackPathPrefix for ANNOUNCE_PLEASE message", slog.String("error", err.Error()))
		return n, err
	}

	aim.Parameters, _, err = ReadParameters(mr)
	if err != nil {
		slog.Error("failed to read Parameters for ANNOUNCE_PLEASE message", slog.String("error", err.Error()))
		return n, err
	}

	slog.Debug("decoded an ANNOUNCE_PLEASE message")

	return n, nil
}
