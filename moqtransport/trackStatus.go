package moqtransport

import (
	"github.com/quic-go/quic-go/quicvarint"
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

func (tsr TrackStatusRequest) serialize() []byte {
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

func (tsr *TrackStatusRequest) deserializeBody(r quicvarint.Reader) error {
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

type TrackStatusMessage struct {
	/*
	 * Track namespace
	 */
	TrackNamespace string

	/*
	 * Track name
	 */
	TrackName string

	/*
	 * Status code
	 */
	Code         TrackStatusCode
	LastGroupID  groupID
	LastObjectID objectID
}

func (ts TrackStatusMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * TRACK_STATUS Message {
	 *   Track Namespace ([]byte),
	 *   Track Name ([]byte),
	 *   Status Code (varint),
	 *   Last Group ID (varint),
	 *   Last Object ID (varint),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */

	// Append the type of the message
	b = quicvarint.Append(b, uint64(TRACK_STATUS))

	// Append Track Namespace
	b = quicvarint.Append(b, uint64(len(ts.TrackNamespace)))
	b = append(b, []byte(ts.TrackNamespace)...)

	// Append Track Name
	b = quicvarint.Append(b, uint64(len(ts.TrackName)))
	b = append(b, []byte(ts.TrackName)...)

	// Append Status Code
	b = quicvarint.Append(b, uint64(ts.Code))

	// Last Group ID
	b = quicvarint.Append(b, uint64(ts.LastGroupID))

	// Last Object ID
	b = quicvarint.Append(b, uint64(ts.LastObjectID))

	return b
}

func (ts *TrackStatusMessage) deserializeBody(r quicvarint.Reader) error {
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
	ts.TrackNamespace = string(buf)

	// Get length of the Track Name
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get Track Name
	buf = make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	ts.TrackName = string(buf)

	// Get Status Code
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ts.Code = TrackStatusCode(num)

	// Get Last Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ts.LastGroupID = groupID(num)

	// Get Last Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ts.LastObjectID = objectID(num)

	return nil
}
