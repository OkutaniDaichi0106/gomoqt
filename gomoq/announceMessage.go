package gomoq

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceMessage struct {
	/*
	 * Track Namespace
	 */
	TrackNamespace string

	/*
	 * Announce Parameters
	 * Parameters should include track authorization information
	 */
	Parameters Parameters
}

func (a AnnounceMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * ANNOUNCE Payload {
	 *   Track Namespace ([]byte),
	 *   Number of Parameters (),
	 *   Announce Parameters(..)
	 * }
	 */

	// TODO?: Chech track namespace exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(ANNOUNCE))
	// Append the supported versions
	b = quicvarint.Append(b, uint64(len(a.TrackNamespace)))
	b = append(b, []byte(a.TrackNamespace)...)

	// Serialize the parameters and append it
	/*
	 * Announce Parameters {
	 *   [Authorization Info Parameter (stirng)],
	 *   [Optional Patameters(..)],
	 * }
	 */
	b = a.Parameters.append(b)

	return b
}

func (a *AnnounceMessage) deserialize(r quicvarint.Reader) error {
	// Get Message ID and check it
	id, err := deserializeHeader(r)
	if err != nil {
		return err
	}
	if id != ANNOUNCE {
		return errors.New("unexpected message")
	}

	return a.deserializeBody(r)
}

func (a *AnnounceMessage) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get length of the track namespace
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

	a.TrackNamespace = string(buf)

	err = a.Parameters.parse(r)
	if err != nil {
		return err
	}

	return nil
}
