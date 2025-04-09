package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
	"golang.org/x/exp/slog"
)

type InfoRequestMessage struct {
	/*
	 * Track name
	 */
	TrackPath string
}

func (irm InfoRequestMessage) Len() int {
	return stringLen(irm.TrackPath)
}

func (irm InfoRequestMessage) Encode(w io.Writer) (int, error) {
	p := GetBytes()
	defer PutBytes(p)

	p = AppendNumber(p, uint64(irm.Len()))
	p = AppendString(p, irm.TrackPath)

	n, err := w.Write(p)
	if err != nil {
		slog.Error("failed to write INFO_REQUEST message", "error", err)
		return n, err
	}

	slog.Debug("encoded an INFO_REQUEST message")

	return n, nil
}

func (irm *InfoRequestMessage) Decode(r io.Reader) (int, error) {

	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read payload for INFO_REQUEST message", "error", err)
		return n, err
	}

	mr := bytes.NewReader(buf)

	irm.TrackPath, _, err = ReadString(mr)
	if err != nil {
		return n, err
	}

	slog.Debug("decoded an INFO_REQUEST message")

	return n, nil
}
