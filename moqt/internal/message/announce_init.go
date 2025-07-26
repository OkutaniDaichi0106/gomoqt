package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
 *	ANNOUNCE_INIT Message {
 *	  Track Pattern (string),
 *	}
 */
type AnnounceInitMessage struct {
	Suffixes []string
}

func (aim AnnounceInitMessage) Len() int {
	var l int

	l += quicvarint.Len(uint64(len(aim.Suffixes)))

	for _, suffix := range aim.Suffixes {
		l += quicvarint.Len(uint64(len(suffix))) + len(suffix)
	}

	return l
}

func (aim AnnounceInitMessage) Encode(dst io.Writer) error {
	msgLen := aim.Len()
	b := pool.Get(msgLen)

	b = quicvarint.Append(b, uint64(msgLen))
	b = quicvarint.Append(b, uint64(len(aim.Suffixes)))

	for _, suffix := range aim.Suffixes {
		b = quicvarint.Append(b, uint64(len(suffix)))
		b = append(b, suffix...)
	}

	_, err := dst.Write(b)

	pool.Put(b)
	return err
}

func (aim *AnnounceInitMessage) Decode(src io.Reader) error {
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

	count, err := ReadVarint(r)
	if err != nil {
		pool.Put(b)
		return err
	}

	aim.Suffixes = make([]string, count)
	var str string
	for i := range count {
		str, err = ReadString(r)
		if err != nil {
			pool.Put(b)
			return err
		}
		aim.Suffixes[i] = str
	}

	pool.Put(b)

	return nil
}
