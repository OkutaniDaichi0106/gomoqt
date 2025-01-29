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

func (irm InfoRequestMessage) Len() int {
	return stringArrayLen(irm.TrackPath)
}

func (irm InfoRequestMessage) Encode(w io.Writer) (int, error) {
	p := GetBytes()
	defer PutBytes(p)

	*p = AppendNumber(*p, uint64(irm.Len()))
	*p = AppendStringArray(*p, irm.TrackPath)

	return w.Write(*p)
}

func (irm *InfoRequestMessage) Decode(r io.Reader) (int, error) {
	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	mr := bytes.NewReader(buf)

	irm.TrackPath, _, err = ReadStringArray(mr)
	if err != nil {
		return n, err
	}

	return n, nil
}
