package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type reader interface {
	quicvarint.Reader
}

func newReader(r io.Reader) (reader, error) {
	// Get a message reader
	num, err := quicvarint.Read(quicvarint.NewReader(r))
	if err != nil {
		slog.Error("failed to get a new message reader", slog.String("error", err.Error()))
		return nil, err
	}

	reader := io.LimitReader(r, int64(num))

	return quicvarint.NewReader(reader), nil
}
