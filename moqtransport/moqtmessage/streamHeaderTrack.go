package moqtmessage

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

func DeserializeStreamHeader(r quicvarint.Reader) (StreamHeader, error) {
	id, err := DeserializeStreamTypeID(r)
	if err != nil {
		return nil, err
	}

	switch id {
	case DATAGRAM_ID:
		var header StreamHeaderDatagram
		err := header.DeserializeStreamHeaderBody(r)
		if err != nil {
			return nil, err
		}

		return &header, nil
	case TRACK_ID:
		var header StreamHeaderTrack
		err := header.DeserializeStreamHeaderBody(r)
		if err != nil {
			return nil, err
		}

		return &header, nil
	case PEEP_ID:
		var header StreamHeaderPeep
		err := header.DeserializeStreamHeaderBody(r)
		if err != nil {
			return nil, err
		}

		return &header, nil
	default:
		return nil, errors.New("unexpected stream type")
	}
}

type StreamHeader interface {
	Serialize() []byte
	DeserializeStreamHeaderBody(quicvarint.Reader) error
	ForwardingPreference() ObjectForwardingPreference
	GetSubscribeID() SubscribeID
	// TrackAlias() TrackAlias
}

/*
 * Deserialize the Message ID
 */
func DeserializeStreamTypeID(r quicvarint.Reader) (StreamTypeID, error) {
	// Get the first number expected to be Stream Type
	num, err := quicvarint.Read(r)
	if err != nil {
		return 0, err
	}
	switch StreamTypeID(num) {
	case DATAGRAM_ID,
		TRACK_ID,
		PEEP_ID:
		return StreamTypeID(num), nil
	default:
		return 0, errors.New("undefined Stream Type ID")
	}
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

var _ StreamHeader = (*StreamHeaderTrack)(nil)

func (sht StreamHeaderTrack) Serialize() []byte {
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
	b = quicvarint.Append(b, uint64(TRACK_ID))
	// Append the Subscriber ID
	b = quicvarint.Append(b, uint64(sht.SubscribeID))
	// Append the Track Alias
	b = quicvarint.Append(b, uint64(sht.TrackAlias))
	// Append the Publisher Priority
	b = quicvarint.Append(b, uint64(sht.PublisherPriority))

	return b
}

func (sht *StreamHeaderTrack) DeserializeStreamHeaderBody(r quicvarint.Reader) error {
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

func (sht StreamHeaderTrack) ForwardingPreference() ObjectForwardingPreference {
	return TRACK
}

func (sht *StreamHeaderTrack) GetSubscribeID() SubscribeID {
	return sht.SubscribeID
}

type GroupChunk struct {
	GroupID
	ObjectChunk
}

func (gc GroupChunk) Serialize() []byte {
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

func (gc *GroupChunk) DeserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	gc.GroupID = GroupID(num)

	gc.ObjectChunk.DeserializeBody(r)

	return nil
}
