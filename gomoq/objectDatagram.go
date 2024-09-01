package gomoq

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
 * OBJECT_DATAGRAM is single object in a datagram
 * and must be the only message on the unidirectional stream
 */
type ObjectDatagram struct {
	Object
	SubscribeID
	TrackAlias
}

func (od ObjectDatagram) serialize() []byte {
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

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(OBJECT_STREAM))
	// Append Subscribe ID
	b = quicvarint.Append(b, uint64(od.SubscribeID))
	// Append Track Alias
	b = quicvarint.Append(b, uint64(od.TrackAlias))
	// Append Group ID
	b = quicvarint.Append(b, uint64(od.GroupChunk.GroupID))
	// Append Object ID
	b = quicvarint.Append(b, uint64(od.GroupChunk.ObjectID))
	// Append Object ID
	b = quicvarint.Append(b, uint64(od.PublisherPriority))
	// Append Object ID
	b = quicvarint.Append(b, uint64(od.StatusCode))

	// Append Object Payload
	b = append(b, od.GroupChunk.Payload...)

	return b
}

// func (od *ObjectDatagram) deserialize(r quicvarint.Reader) error {
// 	// Get Message ID and check it
// 	id, err := deserializeHeader(r)
// 	if err != nil {
// 		return err
// 	}
// 	if id != OBJECT_DATAGRAM {
// 		return errors.New("unexpected message")
// 	}

// 	return od.deserializeBody(r)
// }

func (od *ObjectDatagram) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	od.SubscribeID = SubscribeID(num)

	// Get Track Alias
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	od.TrackAlias = TrackAlias(num)

	// TODO?: Get Track Namespace and Track Name from Track Alias

	// Get Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	od.GroupChunk.GroupID = GroupID(num)

	// Get Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	od.GroupChunk.ObjectID = ObjectID(num)

	// Get Publisher Priority
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	od.PublisherPriority = PublisherPriority(num)

	// Get Object Status Code
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	od.StatusCode = ObjectStatusCode(num)

	// Get Object Payload
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	od.GroupChunk.Payload = buf

	return nil
}
