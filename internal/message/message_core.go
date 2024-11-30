package message

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type Reader quicvarint.Reader

func NewReader(r quicvarint.Reader) (Reader, error) {
	// Get a payload reader
	num, err := quicvarint.Read(r)
	if err != nil {
		return nil, err
	}

	reader := io.LimitReader(r, int64(num))

	return quicvarint.NewReader(reader), nil
}
