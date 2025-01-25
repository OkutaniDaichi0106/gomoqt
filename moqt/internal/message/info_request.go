package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type InfoRequestMessage struct {
	/*
	 * Track name
	 */
	TrackPath []string
}

func (irm InfoRequestMessage) Encode(w io.Writer) (int, error) {
	p := make([]byte, 0, 1<<8)
	p = appendStringArray(p, irm.TrackPath)

	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))
	b = appendBytes(b, p)

	return w.Write(b)
}

func (irm *InfoRequestMessage) Decode(r io.Reader) (int, error) {
	buf, n, err := readBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	mr := bytes.NewReader(buf)

	irm.TrackPath, _, err = readStringArray(mr)
	if err != nil {
		return n, err
	}

	return n, nil
}
