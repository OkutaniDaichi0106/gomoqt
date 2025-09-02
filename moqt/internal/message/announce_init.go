package message

import (
	"io"
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
	return StringArrayLen(aim.Suffixes)
}

func (aim AnnounceInitMessage) Encode(dst io.Writer) error {
	msgLen := aim.Len()
	b := pool.Get(msgLen + VarintLen(uint64(msgLen)))
	defer pool.Put(b)

	b, _ = WriteVarint(b, uint64(msgLen))
	b, _ = WriteStringArray(b, aim.Suffixes)

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
