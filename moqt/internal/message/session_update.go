package message

import (
	"bytes"
	"io"

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
		return n, err
	}

	return n, nil
}

func (sum *SessionUpdateMessage) Decode(r io.Reader) (int, error) {

	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := ReadNumber(mr)
	if err != nil {
		return n, err
	}
	sum.Bitrate = num

	return n, nil
}
