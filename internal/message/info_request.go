package message

import (
	"io"
	"log"

	"github.com/quic-go/quic-go/quicvarint"
)

type InfoRequestMessage struct {
	/*
	 * Track namespace
	 */
	TrackNamespace TrackNamespace

	/*
	 * Track name
	 */
	TrackName string
}

func (irm InfoRequestMessage) Encode(w io.Writer) error {
	/*
	 * Serialize the payload in the following format
	 *
	 * TRACK_STATUS_REQUEST Message Payload {
	 *   Track Namespace (tuple),
	 *   Track Name ([]byte),
	 * }
	 */

	p := make([]byte, 0, 1<<8)

	// Append the Track Namespace
	p = appendTrackNamespace(p, irm.TrackNamespace)

	// Append the Track Name
	p = quicvarint.Append(p, uint64(len(irm.TrackName)))
	p = append(p, []byte(irm.TrackName)...)

	log.Print("INFO_REQUEST payload", p)

	// Serialize the whole message
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)

	return err
}

func (irm *InfoRequestMessage) Decode(r io.Reader) error {
	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get a Track Namespace
	tns, err := readTrackNamespace(mr)
	if err != nil {
		return err
	}
	irm.TrackNamespace = tns

	// Get a Track Name
	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	irm.TrackName = string(buf)

	return nil
}
