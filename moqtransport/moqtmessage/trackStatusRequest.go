package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * Track Status
 */
type TrackStatusCode byte

const (
	TRACK_STATUS_IN_PROGRESS   TrackStatusCode = 0x00
	TRACK_STATUS_NOT_EXIST     TrackStatusCode = 0x01
	TRACK_STATUS_NOT_BEGUN_YET TrackStatusCode = 0x02
	TRACK_STATUS_FINISHED      TrackStatusCode = 0x03
	TRACK_STATUS_RELAY         TrackStatusCode = 0x04
)

type TrackStatusRequest struct {
	/*
	 * Track namespace
	 */
	TrackNamespace string
	/*
	 * Track name
	 */
	TrackName string
}

func (tsr TrackStatusRequest) Serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * TRACK_STATUS_REQUEST Message {
	 *   Track Namespace ([]byte),
	 *   Track Name ([]byte),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(TRACK_STATUS_REQUEST))
	// Append Track Namespace
	b = quicvarint.Append(b, uint64(len(tsr.TrackNamespace)))
	b = append(b, []byte(tsr.TrackNamespace)...)
	// Append Track Name
	b = quicvarint.Append(b, uint64(len(tsr.TrackName)))
	b = append(b, []byte(tsr.TrackName)...)

	return b
}

func (tsr *TrackStatusRequest) DeserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64
	// Get length of the Track Namespace
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get Track Namespace
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	tsr.TrackNamespace = string(buf)

	// Get Track Name
	buf = make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	tsr.TrackName = string(buf)

	return nil
}
