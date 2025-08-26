package message

import (
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
	defer pool.Put(b)

	b = quicvarint.Append(b, uint64(msgLen))
	b = quicvarint.Append(b, uint64(len(aim.Suffixes)))

	for _, suffix := range aim.Suffixes {
		b = quicvarint.Append(b, uint64(len(suffix)))
		b = append(b, suffix...)
	}

	_, err := dst.Write(b)

	return err
}

func (aim *AnnounceInitMessage) Decode(src io.Reader) error {
	num, err := ReadMessageLength(src)
	if err != nil {
		return err
	}

	b := pool.Get(int(num))[:num]
	defer pool.Put(b)

	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	count, n, err := ReadVarint(b)
	if err != nil {
		return err
	}
	b = b[n:]

	aim.Suffixes = make([]string, count)
	var str string
	for i := range aim.Suffixes {
		str, n, err = ReadString(b)
		if err != nil {
			return err
		}
		aim.Suffixes[i] = str
		b = b[n:]
	}

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}
