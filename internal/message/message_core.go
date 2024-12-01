package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type Reader interface {
	quicvarint.Reader
}

func NewReader(str io.Reader) (Reader, error) {
	// Get a message reader
	num, err := quicvarint.Read(quicvarint.NewReader(str))
	if err != nil {
		slog.Debug("failed to get a new message reader")
		return nil, err
	}

	reader := io.LimitReader(str, int64(num))

	return quicvarint.NewReader(reader), nil
}
