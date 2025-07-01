package message

import (
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

func (aim AnnouncePleaseMessage) Encode(w io.Writer) error {
	// Serialize the payload
	p := getBytes()
	defer putBytes(p)

	p = AppendString(p, aim.TrackPrefix)

	_, err := w.Write(p)
	return err
}

func (aim *AnnouncePleaseMessage) Decode(r io.Reader) error {
	var err error
	aim.TrackPrefix, _, err = ReadString(quicvarint.NewReader(r))
	if err != nil {
		return err
	}

	return nil
}
