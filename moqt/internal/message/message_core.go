package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type reader interface {
	quicvarint.Reader
}

func newReader(str io.Reader) (reader, error) {
	// Get a message reader
	num, err := quicvarint.Read(quicvarint.NewReader(str))
	if err != nil {
		slog.Error("failed to get a new message reader", slog.String("error", err.Error()))
		return nil, err
	}

	reader := io.LimitReader(str, int64(num))

	return quicvarint.NewReader(reader), nil
}
