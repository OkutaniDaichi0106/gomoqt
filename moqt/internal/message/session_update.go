package message

import (
	"bytes"
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type SessionUpdateMessage struct {
	/*
	 * Versions selected by the server
	 */
	Bitrate uint64
}

func (sum SessionUpdateMessage) Len() int {
	return numberLen(sum.Bitrate)
}

func (sum SessionUpdateMessage) Encode(w io.Writer) (int, error) {
	p := GetBytes()
	defer PutBytes(p)

	p = AppendNumber(p, uint64(sum.Len()))

	p = AppendNumber(p, sum.Bitrate)

	n, err := w.Write(p)
	if err != nil {
		slog.Error("failed to write a SESSION_UPDATE message", "error", err)
		return n, err
	}

	slog.Debug("encoded a SESSION_UPDATE message")

	return n, nil
}

func (sum *SessionUpdateMessage) Decode(r io.Reader) (int, error) {

	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read payload for SESSION_UPDATE message", "error", err)
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read bitrate", "error", err)
		return n, err
	}
	sum.Bitrate = num

	slog.Debug("decoded a SESSION_UPDATE message")

	return n, nil
}
