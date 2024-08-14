package gomoq

import (
	"errors"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

// Group ID
type GroupID uint64

// Object ID
type ObjectID uint64

/*
 * Track Alias
 * This must be more than 1
 * If This is 0, throw an error
 */
type TrackAlias uint64

// Publisher Priority
type PublisherPriority byte

/*
 * Forwarding Preference
 * Following type are defined in the official document
 * TRACK, GROUP, OBJECT, DATAGRAM
 */
type ForwardingPreference uint64

const (
	TRACK ForwardingPreference = iota
	GROUP
	OBJECT
	DATAGRAM
)

// Object Status
type ObjectStatusCode uint64

const (
	NOMAL_OBJECT       ObjectStatusCode = 0x00
	NONEXISTENT_OBJECT ObjectStatusCode = 0x01
	NONEXISTENT_GROUP  ObjectStatusCode = 0x02
	END_OF_GROUP       ObjectStatusCode = 0x03
	END_OF_TRACK       ObjectStatusCode = 0x04
)

// Canonical Object Model
type Object struct {
	/*
	 * Track's namespace
	 */
	TrackNameSpace string

	/*
	 * Track's name
	 */
	TrackName string

	/*
	 * An 8 bit integer indicating the publisher's priority for the Object
	 */
	PublisherPriority

	/*
	 * Forwarding Preference
	 */
	ForwardingPreference

	/*
	 * A number indicating the status of this object
	 * This used to indicate missing objects or mark the end of a group or track
	 */
	StatusCode ObjectStatusCode

	GroupChunk
}

/*
 * OBJECT_STREAM is single object on a unidirectional stream
 * and must be the only message on the unidirectional stream
 */
type ObjectStream struct {
	Object
	SubscribeID
	TrackAlias
}

func (os ObjectStream) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * OBJECT_STREAM Message {
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
	b = quicvarint.Append(b, uint64(os.SubscribeID))
	// Append Track Alias
	b = quicvarint.Append(b, uint64(os.TrackAlias))
	// Append Group ID
	b = quicvarint.Append(b, uint64(os.GroupID))
	// Append Object ID
	b = quicvarint.Append(b, uint64(os.ObjectID))
	// Append Publisher Priority
	b = quicvarint.Append(b, uint64(os.PublisherPriority))
	// Append Object Status Code
	b = quicvarint.Append(b, uint64(os.StatusCode))

	// Append Object Payload
	b = append(b, os.Payload...)

	return b
}

func (os *ObjectStream) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != OBJECT_STREAM {
		return errors.New("unexpected message")
	}

	// Stream
	os.ForwardingPreference = OBJECT

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	os.SubscribeID = SubscribeID(num)

	// Get Track Alias
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	os.TrackAlias = TrackAlias(num)

	// TODO?: Get Track Namespace and Track Name from Track Alias

	// Get Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	os.GroupID = GroupID(num)

	// Get Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	os.ObjectID = ObjectID(num)

	// Get Publisher Priority
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	os.PublisherPriority = PublisherPriority(num)

	// Get Object Status Code
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	os.StatusCode = ObjectStatusCode(num)

	// Get Object Payload
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	os.Payload = buf

	return nil
}

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
	b = quicvarint.Append(b, uint64(od.GroupID))
	// Append Object ID
	b = quicvarint.Append(b, uint64(od.ObjectID))
	// Append Object ID
	b = quicvarint.Append(b, uint64(od.PublisherPriority))
	// Append Object ID
	b = quicvarint.Append(b, uint64(od.StatusCode))

	// Append Object Payload
	b = append(b, od.Payload...)

	return b
}

func (od *ObjectDatagram) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != OBJECT_DATAGRAM {
		return errors.New("unexpected message")
	}

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
	od.GroupID = GroupID(num)

	// Get Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	od.ObjectID = ObjectID(num)

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
	od.Payload = buf

	return nil
}

type StreamHeaderTrack struct {
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
	 * An 8 bit integer indicating the publisher's priority for the object
	 */
	PublisherPriority
}

func (sht StreamHeaderTrack) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * STREAM_HEADER_TRACK Message {
	 *   Subscribe ID (varint),
	 *   Track Alias (varint),
	 *   Publisher Priority (8),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(STREAM_HEADER_TRACK))
	// Append the Subscriber ID
	b = quicvarint.Append(b, uint64(sht.SubscribeID))
	// Append the Track Alias
	b = quicvarint.Append(b, uint64(sht.TrackAlias))
	// Append the Publisher Priority
	b = quicvarint.Append(b, uint64(sht.PublisherPriority))

	return b
}

func (sht *StreamHeaderTrack) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != STREAM_HEADER_TRACK {
		return errors.New("unexpected message")
	}

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	sht.SubscribeID = SubscribeID(num)

	// Get Track Alias
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	sht.TrackAlias = TrackAlias(num)

	// Get Publisher Priority
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if num >= 1<<8 {
		return errors.New("publiser priority is not an 8 bit integer")
	}
	sht.PublisherPriority = PublisherPriority(num)

	return nil
}

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

func (shg *StreamHeaderGroup) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Message ID and check it
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if MessageID(num) != STREAM_HEADER_GROUP {
		return errors.New("unexpected message")
	}

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

type GroupChunk struct {
	GroupID
	ObjectChunk
}

func (gc GroupChunk) serialize() []byte {
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
	b = quicvarint.Append(b, uint64(gc.GroupID))

	// Append Subscribe ID
	b = quicvarint.Append(b, uint64(gc.ObjectID))

	// Append length of the Payload
	b = quicvarint.Append(b, uint64(len(gc.Payload)))

	// Append Object Payload
	b = append(b, gc.Payload...)

	return b
}

func (gc *GroupChunk) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	gc.GroupID = GroupID(num)

	// Get Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	gc.ObjectID = ObjectID(num)

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
	gc.Payload = buf

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
