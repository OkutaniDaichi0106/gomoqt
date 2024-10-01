package moqtmessage

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

var _ StreamHeader = (*StreamHeaderDatagram)(nil)

type StreamHeaderDatagram struct {
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

	groupID GroupID

	objectID ObjectID
}

func (sht StreamHeaderDatagram) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
	 *
	 * STREAM_HEADER_TRACK Message {
	 *   Subscribe ID (varint),
	 *   Track Alias (varint),
	 *   Publisher Priority (8),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<6) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(DATAGRAM_ID))
	// Append the Subscriber ID
	b = quicvarint.Append(b, uint64(sht.SubscribeID))
	// Append the Track Alias
	b = quicvarint.Append(b, uint64(sht.TrackAlias))
	// Append the Publisher Priority
	b = quicvarint.Append(b, uint64(sht.PublisherPriority))

	return b
}

func (sht *StreamHeaderDatagram) DeserializeStreamHeaderBody(r quicvarint.Reader) error {
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

func (sht StreamHeaderDatagram) ForwardingPreference() ObjectForwardingPreference {
	return DATAGRAM
}

func (shd StreamHeaderDatagram) GetSubscribeID() SubscribeID {
	return shd.SubscribeID
}

func (shd *StreamHeaderDatagram) GetTrackAlias() TrackAlias {
	return shd.TrackAlias
}

func (shd *StreamHeaderDatagram) GetPublisherPriority() PublisherPriority {
	return shd.PublisherPriority
}

func (shd StreamHeaderDatagram) NextGroupHeader() StreamHeaderDatagram {
	return StreamHeaderDatagram{
		SubscribeID:       shd.SubscribeID,
		TrackAlias:        shd.TrackAlias,
		PublisherPriority: shd.PublisherPriority,
		groupID:           shd.groupID + 1,
		objectID:          0,
	}
}

func (shd StreamHeaderDatagram) NextObjectHeader() StreamHeaderDatagram {
	return StreamHeaderDatagram{
		SubscribeID:       shd.SubscribeID,
		TrackAlias:        shd.TrackAlias,
		PublisherPriority: shd.PublisherPriority,
		groupID:           shd.groupID,
		objectID:          shd.objectID + 1,
	}
}
