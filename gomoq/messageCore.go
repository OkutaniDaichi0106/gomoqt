package gomoq

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

// Moqt messages
type MessageID byte

const (
	OBJECT_STREAM        MessageID = 0x00
	OBJECT_DATAGRAM      MessageID = 0x01
	SUBSCRIBE_UPDATE     MessageID = 0x02
	SUBSCRIBE            MessageID = 0x03
	SUBSCRIBE_OK         MessageID = 0x04
	SUBSCRIBE_ERROR      MessageID = 0x05
	ANNOUNCE             MessageID = 0x06
	ANNOUNCE_OK          MessageID = 0x07
	ANNOUNCE_ERROR       MessageID = 0x08
	UNANNOUNCE           MessageID = 0x09
	UNSUBSCRIBE          MessageID = 0x0A
	SUBSCRIBE_DONE       MessageID = 0x0B
	ANNOUNCE_CANCEL      MessageID = 0x0C
	TRACK_STATUS_REQUEST MessageID = 0x0D
	TRACK_STATUS         MessageID = 0x0E
	GOAWAY               MessageID = 0x10
	CLIENT_SETUP         MessageID = 0x40
	SERVER_SETUP         MessageID = 0x41
	STREAM_HEADER_TRACK  MessageID = 0x50
	STREAM_HEADER_GROUP  MessageID = 0x51
)

type Messager interface {
	/*
	 * Serialize values of the fields to byte string
	 * and returns it
	 */
	serialize() []byte

	/*
	 * Deserialize byte string to a value
	 * and reflect the value to the fields
	 */
	deserialize(quicvarint.Reader) error
}

/*
 * Subscription Filter
 *
 * Following type are defined in the official document
 * LATEST_GROUP
 * LATEST_OBJECT
 * ABSOLUTE_START
 * ABSOLUTE_RANGE
 */
type FilterCode uint64

type SubscriptionFilter struct {
	/*
	 * Filter FilterCode indicates the type of filter
	 * This indicates whether the StartGroup/StartObject and EndGroup/EndObject fields
	 * will be present
	 */
	FilterCode

	/*
	 * StartGroupID used only for "AbsoluteStart" or "AbsoluteRange"
	 */
	startGroup GroupID

	/*
	 * StartObjectID used only for "AbsoluteStart" or "AbsoluteRange"
	 */
	startObject ObjectID

	/*
	 * EndGroupID used only for "AbsoluteRange"
	 */
	endGroup GroupID

	/*
	 * EndObjectID used only for "AbsoluteRange".
	 * When it is 0, it means the entire group is required
	 */
	endObject ObjectID
}

const (
	LATEST_GROUP   FilterCode = 0x01
	LATEST_OBJECT  FilterCode = 0x02
	ABSOLUTE_START FilterCode = 0x03
	ABSOLUTE_RANGE FilterCode = 0x04
)

func (sf SubscriptionFilter) isOK() bool {
	//TODO: Check if the Filter Code is valid and valid parameters is set
	if sf.FilterCode == LATEST_GROUP {

	} else if sf.FilterCode == LATEST_OBJECT {

	} else if sf.FilterCode == ABSOLUTE_START {

	} else if sf.FilterCode == ABSOLUTE_RANGE {
		// Check if the Start Group ID is smaller than End Group ID
		if sf.startGroup > sf.endGroup {
			return false
		}
	} else {
		return false
	}
	return true
}

func (sf SubscriptionFilter) append(b []byte) []byte {
	if sf.FilterCode == LATEST_GROUP {
		b = quicvarint.Append(b, uint64(sf.FilterCode))
	} else if sf.FilterCode == LATEST_OBJECT {
		b = quicvarint.Append(b, uint64(sf.FilterCode))
	} else if sf.FilterCode == ABSOLUTE_START {
		// Append Filter Type, Start Group ID and Start Object ID
		b = quicvarint.Append(b, uint64(sf.FilterCode))
		b = quicvarint.Append(b, uint64(sf.startGroup))
		b = quicvarint.Append(b, uint64(sf.startObject))
	} else if sf.FilterCode == ABSOLUTE_RANGE {
		// Append Filter Type, Start Group ID, Start Object ID, End Group ID and End Object ID
		b = quicvarint.Append(b, uint64(sf.FilterCode))
		b = quicvarint.Append(b, uint64(sf.startGroup))
		b = quicvarint.Append(b, uint64(sf.startObject))
		b = quicvarint.Append(b, uint64(sf.endGroup))
		b = quicvarint.Append(b, uint64(sf.endObject))
	} else {
		panic("invalid filter")
	}
	return b
}

type GoAwayMessage struct {
	/*
	 * New session URI
	 * If this is 0 byte, this should be set to current session URI
	 */
	NewSessionURI string
}

func (ga GoAwayMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * GOAWAY Payload {
	 *   New Session URI ([]byte),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(GOAWAY))
	// Append the supported versions
	b = quicvarint.Append(b, uint64(len(ga.NewSessionURI)))
	b = append(b, []byte(ga.NewSessionURI)...)

	return b
}

func (ga *GoAwayMessage) deserialize(r quicvarint.Reader) error {

	// Get Message ID and check it
	id, err := deserializeHeader(r)
	if err != nil {
		return err
	}
	if id != GOAWAY { //TODO: this would means protocol violation
		return errors.New("unexpected message")
	}

	return ga.deserializeBody(r)
}
func (ga *GoAwayMessage) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get length of the URI
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// if num == 0 {
	// 	// TODO: Reuse currenct URI
	// }

	// Get URI
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	ga.NewSessionURI = string(buf)

	// Just one URI supposed to be detected
	// Over one URI will not be detected

	return nil
}

func deserializeHeader(r quicvarint.Reader) (MessageID, error) {
	// Get the first number in the message expected to be MessageID
	num, err := quicvarint.Read(r)
	if err != nil {
		return 0xff, err
	}
	switch MessageID(num) {
	case SUBSCRIBE_UPDATE,
		SUBSCRIBE,
		SUBSCRIBE_OK,
		SUBSCRIBE_ERROR,
		ANNOUNCE,
		ANNOUNCE_OK,
		ANNOUNCE_ERROR,
		UNANNOUNCE,
		UNSUBSCRIBE,
		SUBSCRIBE_DONE,
		ANNOUNCE_CANCEL,
		TRACK_STATUS_REQUEST,
		TRACK_STATUS,
		GOAWAY,
		CLIENT_SETUP,
		SERVER_SETUP,
		STREAM_HEADER_TRACK,
		STREAM_HEADER_GROUP:
		return MessageID(num), nil
	default:
		return 0xff, errors.New("undefined Message ID")
	}
}

/*
func getMessageBody(r quicvarint.Reader, id MessageID) (Messager, error) {
	var msg Messager
	var err error
	switch id {
	case OBJECT_STREAM:
		msg = &ObjectStream{}
	case OBJECT_DATAGRAM:
		msg = &ObjectDatagram{}
	case SUBSCRIBE_UPDATE:
		msg = &SubscribeUpdateMessage{}
	case SUBSCRIBE:
		msg = &SubscribeMessage{}
	case SUBSCRIBE_OK:
		msg = &SubscribeOkMessage{}
	case SUBSCRIBE_ERROR:
		msg = &SubscribeError{}
	case ANNOUNCE:
		msg = &AnnounceMessage{}
	case ANNOUNCE_OK:
		msg = &AnnounceOkMessage{}
	case ANNOUNCE_ERROR:
		msg = &AnnounceError{}
	case UNANNOUNCE:
		msg = &UnannounceMessage{}
	case UNSUBSCRIBE:
		msg = &UnsubscribeMessage{}
	case SUBSCRIBE_DONE:
		msg = &SubscribeDoneMessage{}
	case ANNOUNCE_CANCEL:
		msg = &AnnounceCancelMessage{}
	case TRACK_STATUS_REQUEST:
		msg = &TrackStatusRequest{}
	case TRACK_STATUS:
		msg = &TrackStatusMessage{}
	case GOAWAY:
		msg = &GoAwayMessage{}
	case CLIENT_SETUP:
		msg = &ClientSetupMessage{}
	case SERVER_SETUP:
		msg = &ServerSetupMessage{}
	case STREAM_HEADER_TRACK:
		msg = &StreamHeaderTrack{}
	case STREAM_HEADER_GROUP:
		msg = &StreamHeaderGroup{}
	default:
		// If the massage id is not of the moqt message
		// return an error
		return nil, errors.New("invalid MOQT message type")
	}
	err = msg.deserialize(r)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func getMessage(r io.Reader) (Messager, error) {
	qvReader := quicvarint.NewReader(r)
	id, err := getMessageID(qvReader)
	if err != nil {
		return nil, err
	}
	return getMessageBody(qvReader, id)
}
*/
