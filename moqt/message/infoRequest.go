package message

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

func (irm InfoRequestMessage) SerializePayload() []byte {
	/*
	 * Serialize the payload in the following format
	 *
	 * TRACK_STATUS_REQUEST Message Payload {
	 *   Track Namespace (tuple),
	 *   Track Name ([]byte),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Track Namespace
	p = AppendTrackNamespace(p, irm.TrackNamespace)

	// Append the Track Name
	p = quicvarint.Append(p, uint64(len(irm.TrackName)))
	p = append(p, []byte(irm.TrackName)...)

	return p
}

func (irm *InfoRequestMessage) DeserializePayload(r quicvarint.Reader) error {
	// Get a Track Namespace
	tns, err := ReadTrackNamespace(r)
	if err != nil {
		return err
	}
	irm.TrackNamespace = tns

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
	irm.TrackName = string(buf)

	return nil
}
