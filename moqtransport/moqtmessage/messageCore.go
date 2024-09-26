package moqtmessage

import (
	"errors"
	"io"
	"strings"

	"github.com/quic-go/quic-go/quicvarint"
)

// Moqt messages
type MessageID byte

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
	Serialize() []byte

	/*
	 * Deserialize byte string to a value
	 * and reflect the value to the fields
	 */
	//deserialize(quicvarint.Reader) error

	/*
	 * Deserialize byte string to a value
	 * and reflect the value to the fields
	 */
	DeserializeMessagePayload(quicvarint.Reader) error
}

/*
 * Deserialize the Message ID
 */
func DeserializeMessageID(r quicvarint.Reader) (MessageID, error) {
	// Get the first number expected to be Message ID
	num, err := quicvarint.Read(r)
	if err != nil {
		return 0, err
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
		SERVER_SETUP:
		return MessageID(num), nil
	default:
		return 0, errors.New("undefined Message ID")
	}
}

func NewTrackNamespace(values ...string) TrackNamespace {
	return values
}

type TrackNamespace []string

func (tns TrackNamespace) Append(b []byte) []byte {
	// Append the number of the elements of the Track Namespace
	b = quicvarint.Append(b, uint64(len(tns)))

	for _, v := range tns {
		// Append the length of the data
		b = quicvarint.Append(b, uint64(len(v)))

		// Append the data
		b = append(b, []byte(v)...)
	}

	return b
}

func (tns TrackNamespace) Deserialize(r quicvarint.Reader) error {
	// Get the number of the elements of the track namespace
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}

	for i := uint64(0); i < num; i++ {
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}

		buf := make([]byte, num)
		_, err = r.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		tns = append(tns, string(buf))
	}

	return nil
}

func (tns TrackNamespace) GetFullName() string {
	return strings.Join(tns, "")
}

type TrackNamespacePrefix []string

func (tns TrackNamespacePrefix) Append(b []byte) []byte {
	// Append the number of the elements of the Track Namespace
	b = quicvarint.Append(b, uint64(len(tns)))

	for _, v := range tns {
		// Append the length of the data
		b = quicvarint.Append(b, uint64(len(v)))

		// Append the data
		b = append(b, []byte(v)...)
	}

	return b
}

func (tns TrackNamespacePrefix) Deserialize(r quicvarint.Reader) error {
	// Get the number of the elements of the track namespace
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}

	for i := uint64(0); i < num; i++ {
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}

		buf := make([]byte, num)
		_, err = r.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		tns = append(tns, string(buf))
	}

	return nil
}
