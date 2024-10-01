package moqtmessage

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

var _ StreamHeader = (*StreamHeaderPeep)(nil)

type StreamHeaderPeep struct {
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
	 * Peep ID
	 */
	PeepID

	/*
	 * An 8 bit integer indicating the publisher's priority for the object
	 */
	PublisherPriority
}

func (shp StreamHeaderPeep) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * STREAM_HEADER_Peep Message {
	 *   Subscribe ID (varint),
	 *   Track Alias (varint),
	 *   Peep ID (varint),
	 *   Publisher Priority (8),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(PEEP_ID))
	// Append the Subscriber ID
	b = quicvarint.Append(b, uint64(shp.SubscribeID))
	// Append the Track Alias
	b = quicvarint.Append(b, uint64(shp.TrackAlias))
	// Append the Peep ID
	b = quicvarint.Append(b, uint64(shp.PeepID))
	// Append the Publisher Priority
	b = quicvarint.Append(b, uint64(shp.PublisherPriority))

	return b
}

func (shp *StreamHeaderPeep) DeserializeStreamHeaderBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	shp.SubscribeID = SubscribeID(num)

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	shp.TrackAlias = TrackAlias(num)

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	shp.PeepID = PeepID(num)

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if num >= 1<<8 {
		return errors.New("publiser priority is not an 8 bit integer")
	}
	shp.PublisherPriority = PublisherPriority(num)

	return nil
}

func (StreamHeaderPeep) ForwardingPreference() ObjectForwardingPreference {
	return PEEP
}

func (shp *StreamHeaderPeep) GetSubscribeID() SubscribeID {
	return shp.SubscribeID
}

func (shp StreamHeaderPeep) NextGroupHeader() StreamHeaderPeep {
	return StreamHeaderPeep{
		SubscribeID:       shp.SubscribeID,
		TrackAlias:        shp.TrackAlias,
		PublisherPriority: shp.PublisherPriority,
		GroupID:           shp.GroupID + 1,
		PeepID:            0,
	}
}

func (shp StreamHeaderPeep) NextPeepHeader() StreamHeaderPeep {
	return StreamHeaderPeep{
		SubscribeID:       shp.SubscribeID,
		TrackAlias:        shp.TrackAlias,
		PublisherPriority: shp.PublisherPriority,
		GroupID:           shp.GroupID,
		PeepID:            shp.PeepID + 1,
	}
}

type ObjectChunk struct {
	/*
	 * Object ID
	 */
	ObjectID

	Payload []byte

	/*
	 * A number indicating the status of this object
	 * This is only sent if the length of the Object Payload is zero
	 * This indicates missing objects or mark the end of a group or track
	 */
	StatusCode ObjectStatusCode
}

func (oc ObjectChunk) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
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

	// Append the Subscribe ID
	b = quicvarint.Append(b, uint64(oc.ObjectID))

	// Append the length of the Payload
	b = quicvarint.Append(b, uint64(len(oc.Payload)))

	// Append the Object Payload
	b = append(b, oc.Payload...)

	// Append the Object Status if the length of the Object Payload is zero
	if len(oc.Payload) == 0 {
		b = quicvarint.Append(b, uint64(oc.StatusCode))
	}

	return b
}

func (oc *ObjectChunk) DeserializeBody(r quicvarint.Reader) error {
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

	if num > 0 {
		// Get Object Payload
		buf := make([]byte, num)
		_, err = r.Read(buf)
		if err != nil {
			return err
		}
		oc.Payload = buf
	} else if num == 0 {
		// Get Object Status if the length of the Object Payload is zero
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		oc.StatusCode = ObjectStatusCode(num)
	}

	return nil
}

func (shp *StreamHeaderPeep) GetTrackAlias() TrackAlias {
	return shp.TrackAlias
}
