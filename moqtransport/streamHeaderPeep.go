package moqtransport

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type StreamHeaderPeep struct {
	/*
	 * A number to identify the subscribe session
	 */
	subscribeID

	/*
	 * An number indicates a track
	 * This is referenced instead of the Track Name and Track Namespace
	 */
	TrackAlias

	/*
	 * Group ID
	 */
	groupID

	/*
	 * Peep ID
	 */
	peepID

	/*
	 * An 8 bit integer indicating the publisher's priority for the object
	 */
	PublisherPriority
}

func (shg StreamHeaderPeep) serialize() []byte {
	/*
	 * Serialize as following formatt
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
	b = quicvarint.Append(b, uint64(STREAM_HEADER_PEEP))
	// Append the Subscriber ID
	b = quicvarint.Append(b, uint64(shg.subscribeID))
	// Append the Track Alias
	b = quicvarint.Append(b, uint64(shg.TrackAlias))
	// Append the Peep ID
	b = quicvarint.Append(b, uint64(shg.peepID))
	// Append the Publisher Priority
	b = quicvarint.Append(b, uint64(shg.PublisherPriority))

	return b
}

func (shg *StreamHeaderPeep) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	shg.subscribeID = subscribeID(num)

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
	shg.peepID = peepID(num)

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

func (sht StreamHeaderPeep) ForwardingPreference() ForwardingPreference {
	return PEEP
}

type ObjectChunk struct {
	/*
	 * Object ID
	 */
	objectID

	Payload []byte

	/*
	 * A number indicating the status of this object
	 * This is only sent if the length of the Object Payload is zero
	 * This indicates missing objects or mark the end of a group or track
	 */
	StatusCode ObjectStatusCode
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
	b = quicvarint.Append(b, uint64(oc.objectID))

	// Append length of the Payload
	b = quicvarint.Append(b, uint64(len(oc.Payload)))

	// Append Object Payload
	b = append(b, oc.Payload...)

	// Append Object Status if the length of the Object Payload is zero
	if len(oc.Payload) == 0 {
		b = quicvarint.Append(b, uint64(oc.StatusCode))
	}

	return b
}

func (oc *ObjectChunk) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	oc.objectID = objectID(num)

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
