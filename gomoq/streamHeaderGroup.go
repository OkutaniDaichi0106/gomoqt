package gomoq

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type StreamHeaderGroup struct {
	/*
	 * A number to identify the subscribe session
	 */
	SubscribeID

	/*
	 * An number indicates a track
	 * This is referenced instead of the Track Name and Track Namespace
	 */
	TrackAlias
	/*
	 * Group ID
	 */
	GroupID

	/*
	 * An 8 bit integer indicating the publisher's priority for the object
	 */
	PublisherPriority
}

func (shg StreamHeaderGroup) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * STREAM_HEADER_GROUP Message {
	 *   Subscribe ID (varint),
	 *   Track Alias (varint),
	 *   Group ID (varint),
	 *   Publisher Priority (8),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(STREAM_HEADER_GROUP))
	// Append the Subscriber ID
	b = quicvarint.Append(b, uint64(shg.SubscribeID))
	// Append the Track Alias
	b = quicvarint.Append(b, uint64(shg.TrackAlias))
	// Append the Group ID
	b = quicvarint.Append(b, uint64(shg.GroupID))
	// Append the Publisher Priority
	b = quicvarint.Append(b, uint64(shg.PublisherPriority))

	return b
}

// func (shg *StreamHeaderGroup) deserialize(r quicvarint.Reader) error {
// 	// Get Message ID and check it
// 	id, err := deserializeHeader(r)
// 	if err != nil {
// 		return err
// 	}
// 	if id != STREAM_HEADER_GROUP {
// 		return ErrUnexpectedMessage
// 	}

// 	return shg.deserializeBody(r)
// }

func (shg *StreamHeaderGroup) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	shg.SubscribeID = SubscribeID(num)

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	shg.TrackAlias = TrackAlias(num)

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	shg.GroupID = GroupID(num)

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if num >= 1<<8 {
		return errors.New("publiser priority is not an 8 bit integer")
	}
	shg.PublisherPriority = PublisherPriority(num)

	return nil
}

type ObjectChunk struct {
	ObjectID
	Payload []byte
}

func (oc ObjectChunk) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * OBJECT Chunk {
	 *   Object ID (varint),
	 *   Object Status (varint),
	 *   Object Payload (..),
	 *}
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */

	// Append Subscribe ID
	b = quicvarint.Append(b, uint64(oc.ObjectID))

	// Append length of the Payload
	b = quicvarint.Append(b, uint64(len(oc.Payload)))

	// Append Object Payload
	b = append(b, oc.Payload...)

	return b
}

func (oc *ObjectChunk) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	oc.ObjectID = ObjectID(num)

	// Get length of the Object Payload
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get Object Payload
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	oc.Payload = buf

	return nil
}
