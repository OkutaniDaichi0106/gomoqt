package gomoq

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceOkMessage struct {
	TrackNamespace string
}

func (ao AnnounceOkMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * ANNOUNCE_OK Payload {
	 *   Track Namespace ([]byte),
	 * }
	 */

	// TODO?: Chech track namespace exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(ANNOUNCE_OK))
	// Append the supported versions
	b = quicvarint.Append(b, uint64(len(ao.TrackNamespace)))
	b = append(b, []byte(ao.TrackNamespace)...)

	return b
}

func (ao *AnnounceOkMessage) deserialize(r quicvarint.Reader) error {
	// Get Message ID and check it
	id, err := deserializeHeader(r)
	if err != nil {
		return err
	}
	if id != ANNOUNCE_OK { //TODO: this would means protocol violation
		return errors.New("unexpected message")
	}

	return ao.deserializeBody(r)
}

func (ao *AnnounceOkMessage) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get length of the string of the track namespace
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get track namespace
	buf := make([]byte, num)

	_, err = r.Read(buf)
	if err != nil {
		return err
	}

	ao.TrackNamespace = string(buf)

	// Just one track namespace supposed to be detected
	// Over one track namespace will not be detected

	return nil
}
