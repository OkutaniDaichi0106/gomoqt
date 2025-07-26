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
	return VarintLen(sum.Bitrate)
}

func (sum SessionUpdateMessage) Encode(w io.Writer) error {
	msgLen := sum.Len()
	b := pool.Get(msgLen)

	b = quicvarint.Append(b, uint64(msgLen))
	b = quicvarint.Append(b, sum.Bitrate)

	_, err := w.Write(b)
	pool.Put(b)
	return err
}

func (sum *SessionUpdateMessage) Decode(src io.Reader) error {
	num, err := ReadVarint(src)
	if err != nil {
		return err
	}

	b := pool.Get(int(num))[:num]
	_, err = io.ReadFull(src, b)
	if err != nil {
		pool.Put(b)
		return err
	}

	r := bytes.NewReader(b)

	num, err = ReadVarint(r)
	if err != nil {
		pool.Put(b)
		return err
	}
	sum.Bitrate = num

	pool.Put(b)

	return nil
}
