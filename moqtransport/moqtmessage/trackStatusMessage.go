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
	 * Serialize the message in the following formatt
	 *
	 * TRACK_STATUS Message {
	 *   Type (varint) = 0x0E,
	 *   Length (varint),
	 *   Track Namespace (tuple),
	 *   Track Name ([]byte),
	 *   Status Code (varint),
	 *   Last Group ID (varint),
	 *   Last Object ID (varint),
	 * }
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<10)

	// Append the Track Namespace
	p = ts.TrackNamespace.Append(p)

	// Append the Track Name
	p = quicvarint.Append(p, uint64(len(ts.TrackName)))
	p = append(p, []byte(ts.TrackName)...)

	// Append the Status Code
	p = quicvarint.Append(p, uint64(ts.Code))

	// Appen the Last Group ID
	p = quicvarint.Append(p, uint64(ts.LastGroupID))

	// Appen the  Last Object ID
	p = quicvarint.Append(p, uint64(ts.LastObjectID))

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the type of the message
	b = quicvarint.Append(b, uint64(TRACK_STATUS))

	// Appen the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (ts *TrackStatusMessage) DeserializePayload(r quicvarint.Reader) error {
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
