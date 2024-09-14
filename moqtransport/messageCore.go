package moqtransport

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

// Moqt messages
type MessageID byte

const (
	//OBJECT_STREAM       MessageID = 0x00 // Deprecated
	OBJECT_DATAGRAM     MessageID = 0x01
	STREAM_HEADER_TRACK MessageID = 0x50
	STREAM_HEADER_GROUP MessageID = 0x51
	STREAM_HEADER_PEEP  MessageID = 0x52
)

// Control Messages
const (
	SUBSCRIBE_UPDATE          MessageID = 0x02
	SUBSCRIBE                 MessageID = 0x03
	SUBSCRIBE_OK              MessageID = 0x04
	SUBSCRIBE_ERROR           MessageID = 0x05
	ANNOUNCE                  MessageID = 0x06
	ANNOUNCE_OK               MessageID = 0x07
	ANNOUNCE_ERROR            MessageID = 0x08
	UNANNOUNCE                MessageID = 0x09
	UNSUBSCRIBE               MessageID = 0x0A
	SUBSCRIBE_DONE            MessageID = 0x0B
	ANNOUNCE_CANCEL           MessageID = 0x0C
	TRACK_STATUS_REQUEST      MessageID = 0x0D
	TRACK_STATUS              MessageID = 0x0E
	GOAWAY                    MessageID = 0x10
	SUBSCRIBE_NAMESPACE       MessageID = 0x11 //TODO
	SUBSCRIBE_NAMESPACE_OK    MessageID = 0x12 //TODO
	SUBSCRIBE_NAMESPACE_ERROR MessageID = 0x13 //TODO
	UNSUBSCRIBE_NAMESPACE     MessageID = 0x14 //TODO
	CLIENT_SETUP              MessageID = 0x40
	SERVER_SETUP              MessageID = 0x41
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
	//deserialize(quicvarint.Reader) error

	/*
	 * Deserialize byte string to a value
	 * and reflect the value to the fields
	 */
	deserializeBody(quicvarint.Reader) error
}

/*
 * Deserialize the header of the message which is message id
 */
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
		STREAM_HEADER_PEEP:
		return MessageID(num), nil
	default:
		return 0xff, errors.New("undefined Message ID")
	}
}
