package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * Track Status
 */
type TrackStatusCode byte

const (
	TRACK_STATUS_IN_PROGRESS       TrackStatusCode = 0x00
	TRACK_STATUS_NOT_EXIST         TrackStatusCode = 0x01
	TRACK_STATUS_NOT_BEGUN_YET     TrackStatusCode = 0x02
	TRACK_STATUS_FINISHED          TrackStatusCode = 0x03
	TRACK_STATUS_UNTRACEABLE_RELAY TrackStatusCode = 0x04
)

type TrackStatusRequestMessage struct {
	/*
	 * Track namespace
	 */
	TrackNamespace TrackNamespace

	/*
	 * Track name
	 */
	TrackName string

	/*
	 * Track Alias
	 */
	TrackAlias TrackAlias
}

func (tsr TrackStatusRequestMessage) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * TRACK_STATUS_REQUEST Message {
	 *   Track Namespace (tuple),
	 *   Track Name ([]byte),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<10)

	// Append the Track Namespace
	p = AppendTrackNamespace(p, tsr.TrackNamespace)

	// Append the Track Name
	p = quicvarint.Append(p, uint64(len(tsr.TrackName)))
	p = append(p, []byte(tsr.TrackName)...)

	// Append the Track Alias
	p = quicvarint.Append(p, uint64(tsr.TrackAlias))

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the message type
	p = quicvarint.Append(p, uint64(TRACK_STATUS_REQUEST))

	// Appen the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (tsr *TrackStatusRequestMessage) DeserializePayload(r quicvarint.Reader) error {
	// Get a Track Namespace
	tns, err := ReadTrackNamespace(r)
	if err != nil {
		return err
	}
	tsr.TrackNamespace = tns

	// Get a Track Name
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	tsr.TrackName = string(buf)

	// Get a Track Alias
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	tsr.TrackAlias = TrackAlias(num)

	return nil
}
