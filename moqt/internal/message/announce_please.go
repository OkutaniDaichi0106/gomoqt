package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
 *	ANNOUNCE_PLEASE Message {
 *	  Track Prefix (string),
 *	}
 */
type AnnouncePleaseMessage struct {
	TrackPrefix string
}

func (aim AnnouncePleaseMessage) Len() int {
	var l int

	l += StringLen(aim.TrackPrefix)

	return l
}

func (aim AnnouncePleaseMessage) Encode(w io.Writer) error {
	msgLen := aim.Len()
	b := pool.Get(msgLen + quicvarint.Len(uint64(msgLen)))

	b = quicvarint.Append(b, uint64(msgLen))
	b = quicvarint.Append(b, uint64(len(aim.TrackPrefix)))
	b = append(b, aim.TrackPrefix...)

	_, err := w.Write(b)

	pool.Put(b)

	return err
}

func (aim *AnnouncePleaseMessage) Decode(src io.Reader) error {
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

	str, err := ReadString(r)
	if err != nil {
		pool.Put(b)
		return err
	}
	aim.TrackPrefix = str

	pool.Put(b)

	return nil
}
