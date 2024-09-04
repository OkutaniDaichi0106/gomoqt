package gomoq

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type StreamHeader interface {
	Messager
	StreamType() ForwardingPreference
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

func (sht *StreamHeaderTrack) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

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

func (sht StreamHeaderTrack) StreamType() ForwardingPreference {
	return TRACK
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

	if len(gc.Payload) == 0 {
		b = quicvarint.Append(b, uint64(gc.StatusCode))
	}

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
	if num > 0 {
		// Get Object Payload
		buf := make([]byte, num)
		_, err = r.Read(buf)
		if err != nil {
			return err
		}
		gc.Payload = buf
	} else if num == 0 {
		// Get Object Status Code
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		gc.StatusCode = ObjectStatusCode(num)
	}

	return nil
}
