package moqtmessage

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type TrackStatusMessage struct {
	/*
	 * Track namespace
	 */
	TrackNamespace TrackNamespace

	/*
	 * Track name
	 */
	TrackName string

	/*
	 * Status code
	 */
	Code         TrackStatusCode
	LastGroupID  GroupID
	LastObjectID ObjectID
}

func (ts TrackStatusMessage) Serialize() []byte {
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
	b = ts.TrackNamespace.Append(b)

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

func (ts *TrackStatusMessage) DeserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Track Namespace
	var tns TrackNamespace
	err = tns.Deserialize(r)
	if err != nil {
		return err
	}
	ts.TrackNamespace = tns

	// Get length of the Track Name
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get Track Name
	buf := make([]byte, num)
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
	ts.LastGroupID = GroupID(num)

	// Get Last Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ts.LastObjectID = ObjectID(num)

	return nil
}
