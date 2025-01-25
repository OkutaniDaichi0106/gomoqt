package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
 *	ANNOUNCE_PLEASE Message {
 *	  Track Prefix ([]string),
 *	  Announce Parameters (Parameters),
 *	}
 */

type AnnouncePleaseMessage struct {
	TrackPathPrefix []string
	Parameters      Parameters
}

func (aim AnnouncePleaseMessage) Encode(w io.Writer) (int, error) {
	// Serialize the payload
	p := make([]byte, 0, 1<<6) // TODO: Tune the size

	p = appendStringArray(p, aim.TrackPathPrefix)

	p = appendParameters(p, aim.Parameters)

	// Serialize the message
	b := make([]byte, 0, len(p)+quicvarint.Len(uint64(len(p))))

	b = appendBytes(b, p)

	// Write
	return w.Write(b)
}

func (aim *AnnouncePleaseMessage) Decode(r io.Reader) (int, error) {
	// Read the payload
	buf, n, err := readBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	// Decode the payload
	mr := bytes.NewReader(buf)

	aim.TrackPathPrefix, _, err = readStringArray(mr)
	if err != nil {
		return n, err
	}

	aim.Parameters, _, err = readParameters(mr)
	if err != nil {
		return n, err
	}

	return n, nil
}
