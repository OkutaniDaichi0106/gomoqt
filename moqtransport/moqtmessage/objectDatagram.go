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
	 * Serialize the message in the following formatt
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
	// Append the stream type
	b = quicvarint.Append(b, uint64(DATAGRAM_ID))
	// Append the Subscribe ID
	b = quicvarint.Append(b, uint64(od.SubscribeID))
	// Append the Track Alias
	b = quicvarint.Append(b, uint64(od.TrackAlias))
	// Append the Group ID
	b = quicvarint.Append(b, uint64(od.GroupID))
	// Append the Object ID
	b = quicvarint.Append(b, uint64(od.ObjectID))
	// Append the Publisher Priority
	b = quicvarint.Append(b, uint64(od.PublisherPriority))
	// Append the Object Payload Length
	b = quicvarint.Append(b, uint64(len(od.Payload)))
	// Append the Object Payload
	b = append(b, od.Payload...)

	if len(od.Payload) == 0 {
		// Append the Object Status Code
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
