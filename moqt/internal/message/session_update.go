package message

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type SessionUpdateMessage struct {
	/*
	 * Versions selected by the server
	 */
	Bitrate uint64
}

func (sum SessionUpdateMessage) Encode(w io.Writer) error {
	p := getBytes()
	defer putBytes(p)

	p = AppendNumber(p, sum.Bitrate)

	_, err := w.Write(p)
	return err
}

func (sum *SessionUpdateMessage) Decode(r io.Reader) error {
	num, _, err := ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}
	sum.Bitrate = num

	return nil
}
