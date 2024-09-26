package moqtmessage

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * OBJECT_DATAGRAM is single object in a datagram
 * and must be the only message on the unidirectional stream
 */
type ObjectDatagram struct {
	SubscribeID
	TrackAlias
	GroupChunk
	PublisherPriority
}

func (od ObjectDatagram) Serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * OBJECT_DATAGRAM Message {
	 *   Subscribe ID (varint),
	 *   Track Alias (varint),
	 *   Group ID (varint),
	 *   Object ID (varint),
	 *   Publisher Priority (8),
	 *   Object Status (varint),
	 *   Object Payload (..),
	 *}
	 */

	// TODO?: Check URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(DATAGRAM.ID()))
	// Append Subscribe ID
	b = quicvarint.Append(b, uint64(od.SubscribeID))
	// Append Track Alias
	b = quicvarint.Append(b, uint64(od.TrackAlias))
	// Append Group ID
	b = quicvarint.Append(b, uint64(od.GroupID))
	// Append Object ID
	b = quicvarint.Append(b, uint64(od.ObjectID))
	// Append Publisher Priority
	b = quicvarint.Append(b, uint64(od.PublisherPriority))
	// Append Object Payload Length
	b = quicvarint.Append(b, uint64(len(od.Payload)))
	// Append Object Payload
	b = append(b, od.Payload...)

	if len(od.Payload) == 0 {
		// Append Object Status Code
		b = quicvarint.Append(b, uint64(od.StatusCode))
	}

	return b
}

func (od *ObjectDatagram) DeserializeStreamHeader(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get a Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	od.SubscribeID = SubscribeID(num)

	// Get a Track Alias
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	od.TrackAlias = TrackAlias(num)

	// Get a Publisher Priority
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if num >= 1<<8 {
		return errors.New("publiser priority is not an 8 bit integer")
	}
	od.PublisherPriority = PublisherPriority(num)

	return nil
}

func (od *ObjectDatagram) DeserializeBody(r quicvarint.Reader) error {
	return od.GroupChunk.DeserializeBody(r)
}

func (ObjectDatagram) ForwardingPreference() ObjectForwardingPreference {
	return DATAGRAM
}
