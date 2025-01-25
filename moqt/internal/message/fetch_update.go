package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type FetchUpdateMessage struct {
	TrackPriority TrackPriority
}

func (fum FetchUpdateMessage) Encode(w io.Writer) (int, error) {
	p := make([]byte, 0, 1<<3)
	p = appendNumber(p, uint64(fum.TrackPriority))

	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))
	b = appendBytes(b, p)

	return w.Write(b)
}

func (fum *FetchUpdateMessage) Decode(r io.Reader) (int, error) {
	buf, n, err := readBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := readNumber(mr)
	if err != nil {
		return n, err
	}
	fum.TrackPriority = TrackPriority(num)

	return n, nil
}
