package moqtmessage

import (
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

type TrackStatusMessage struct {
	TrackAlias    TrackAlias
	Code          TrackStatusCode
	LatestGroupID GroupID
	GroupOrder    GroupOrder
	GroupExpires  time.Duration
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

	// Append the Track Alias
	p = quicvarint.Append(p, uint64(ts.TrackAlias))

	// Append the Status Code
	p = quicvarint.Append(p, uint64(ts.Code))

	// Appen the Last Group ID
	p = quicvarint.Append(p, uint64(ts.LatestGroupID))

	// Appen the Group Order
	p = quicvarint.Append(p, uint64(ts.GroupOrder))

	// Appen the Group Expires
	p = quicvarint.Append(p, uint64(ts.GroupExpires))

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

	// Get a Track Alias
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ts.TrackAlias = TrackAlias(num)

	// Get a Status Code
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ts.Code = TrackStatusCode(num)

	// Get a Latest Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ts.LatestGroupID = GroupID(num)

	// Get a Group Order
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ts.GroupOrder = GroupOrder(num)

	// Get a Group Expires
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ts.GroupExpires = time.Duration(num)

	return nil
}
