package gomoq

import (
	"errors"
	"io"

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
type SubscriptionFilter uint64

const (
	LATEST_GROUP   SubscriptionFilter = 0x01
	LATEST_OBJECT  SubscriptionFilter = 0x02
	ABSOLUTE_START SubscriptionFilter = 0x03
	ABSOLUTE_RANGE SubscriptionFilter = 0x04
)

type TrackStatusRequest struct {
	/*
	 * Track namespace
	 */
	TrackNamespace string
	/*
	 * Track name
	 */
	TrackName string
}

func (tsr TrackStatusRequest) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * TRACK_STATUS_REQUEST Message {
	 *   Track Namespace ([]byte),
	 *   Track Name ([]byte),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(TRACK_STATUS_REQUEST))
	// Append Track Namespace
	b = quicvarint.Append(b, uint64(len(tsr.TrackNamespace)))
	b = append(b, []byte(tsr.TrackNamespace)...)
	// Append Track Name
	b = quicvarint.Append(b, uint64(len(tsr.TrackName)))
	b = append(b, []byte(tsr.TrackName)...)

	return b
}

func (tsr *TrackStatusRequest) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64
	// Get length of the Track Namespace
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get Track Namespace
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	tsr.TrackNamespace = string(buf)

	// Get Track Name
	buf = make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	tsr.TrackName = string(buf)

	return nil
}

type TrackStatusMessage struct {
	/*
	 * Track namespace
	 */
	TrackNamespace string

	/*
	 * Track name
	 */
	TrackName string

	/*
	 * Status code
	 */
	Code         TrackStatusCode
	LastGroupID  GroupID
	LastObjectID ObjectID
}

func (ts TrackStatusMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * TRACK_STATUS Message {
	 *   Track Namespace ([]byte),
	 *   Track Name ([]byte),
	 *   Status Code (varint),
	 *   Last Group ID (varint),
	 *   Last Object ID (varint),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */

	// Append the type of the message
	b = quicvarint.Append(b, uint64(TRACK_STATUS))

	// Append Track Namespace
	b = quicvarint.Append(b, uint64(len(ts.TrackNamespace)))
	b = append(b, []byte(ts.TrackNamespace)...)

	// Append Track Name
	b = quicvarint.Append(b, uint64(len(ts.TrackName)))
	b = append(b, []byte(ts.TrackName)...)

	// Append Status Code
	b = quicvarint.Append(b, uint64(ts.Code))

	// Last Group ID
	b = quicvarint.Append(b, uint64(ts.LastGroupID))

	// Last Object ID
	b = quicvarint.Append(b, uint64(ts.LastObjectID))

	return b
}

func (ts *TrackStatusMessage) deserialize(r quicvarint.Reader) error {
	var err error
	var num uint64
	// Get length of the Track Namespace
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get Track Namespace
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	ts.TrackNamespace = string(buf)

	// Get length of the Track Name
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	// Get Track Name
	buf = make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	ts.TrackName = string(buf)

	// Get Status Code
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ts.Code = TrackStatusCode(num)

	// Get Last Group ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ts.LastGroupID = GroupID(num)

	// Get Last Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	ts.LastObjectID = ObjectID(num)

	return nil
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

/*
	func readHeader(r quicvarint.Reader) (MessageID, error) {
		var id MessageID
		// Get the first number in the message expected to be MessageID
		firstNum, err := quicvarint.Read(r)
		if err != nil {
			return 0xff, err
		}
		switch MessageID(firstNum) {
		case SUBSCRIBE_UPDATE:
		case SUBSCRIBE:
		case SUBSCRIBE_OK:
		case SUBSCRIBE_ERROR:
		case ANNOUNCE:
		case ANNOUNCE_OK:
		case ANNOUNCE_ERROR:
		case UNANNOUNCE:
		case UNSUBSCRIBE:
		case SUBSCRIBE_DONE:
		case ANNOUNCE_CANCEL:
		case TRACK_STATUS_REQUEST:
		case TRACK_STATUS:
		case GOAWAY:
		case CLIENT_SETUP:
		case SERVER_SETUP:
		case STREAM_HEADER_TRACK:
		case STREAM_HEADER_GROUP:
		default:
			return 0xff, errors.New("undefined Message ID")
		}

		return id, nil
	}
*/
func GetMessage(r io.Reader) (Messager, error) {
	var msg Messager
	// Parse first parameter in a message which is expected to be
	reader := quicvarint.NewReader(r)
	id, err := quicvarint.Read(reader)
	if err != nil {
		return nil, err
	}

	switch MessageID(id) {
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
	err = msg.deserialize(reader)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
