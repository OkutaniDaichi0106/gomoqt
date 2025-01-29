package message

import (
	"bytes"
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type FetchUpdateMessage struct {
	TrackPriority TrackPriority
}

func (fum FetchUpdateMessage) Len() int {
	return numberLen(uint64(fum.TrackPriority))
}

func (fum FetchUpdateMessage) Encode(w io.Writer) (int, error) {
	slog.Debug("encoding a FETCH_UPDATE message")

	p := GetBytes()
	defer PutBytes(p)

	*p = AppendNumber(*p, uint64(fum.Len()))

	*p = AppendNumber(*p, uint64(fum.TrackPriority))

	slog.Debug("encoded a FETCH_UPDATE message")

	return w.Write(*p)
}

func (fum *FetchUpdateMessage) Decode(r io.Reader) (int, error) {
	slog.Debug("decoding a FETCH_UPDATE message")

	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to read bytes for FETCH_UPDATE message", slog.String("error", err.Error()), slog.Int("bytes_read", n))
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := ReadNumber(mr)
	if err != nil {
		slog.Error("failed to read TrackPriority for FETCH_UPDATE message", slog.String("error", err.Error()))
		return n, err
	}
	fum.TrackPriority = TrackPriority(num)

	slog.Debug("decoded a FETCH_UPDATE message", slog.Int("bytes_read", n))

	return n, nil
}
