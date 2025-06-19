package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
 *	ANNOUNCE_PLEASE Message {
 *	  Track Pattern (string),
 *	}
 */
type AnnouncePleaseMessage struct {
	TrackPrefix string
}

func (aim AnnouncePleaseMessage) Len() int {
	// Calculate the length of the payload
	l := 0
	l += stringLen(aim.TrackPrefix)

	return l
}

func (aim AnnouncePleaseMessage) Encode(w io.Writer) (int, error) {
	// Serialize the payload
	p := getBytes()
	defer putBytes(p)

	p = AppendNumber(p, uint64(aim.Len()))

	p = AppendString(p, aim.TrackPrefix)

	return w.Write(p)
}

func (aim *AnnouncePleaseMessage) Decode(r io.Reader) (int, error) {
	// Read the payload
	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	// Decode the payload
	mr := bytes.NewReader(buf)

	aim.TrackPrefix, _, err = ReadString(mr)
	if err != nil {
		return n, err
	}

	return n, nil
}
