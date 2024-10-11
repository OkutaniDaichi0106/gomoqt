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
	SUBSCRIBE_UPDATE     MessageID = 0x02
	SUBSCRIBE            MessageID = 0x03
	ANNOUNCE             MessageID = 0x06
	TRACK_STATUS_REQUEST MessageID = 0x0D
	TRACK_STATUS         MessageID = 0x0E
	GOAWAY               MessageID = 0x10
	SUBSCRIBE_NAMESPACE  MessageID = 0x11
	CLIENT_SETUP         MessageID = 0x40
	SERVER_SETUP         MessageID = 0x41
)

func ReadControlMessage(r quicvarint.Reader) (MessageID, quicvarint.Reader, error) {
	// Get a Message ID
	var messageID MessageID
	num, err := quicvarint.Read(r)
	if err != nil {
		return 0, nil, err
	}
	switch MessageID(num) {
	case SUBSCRIBE_UPDATE,
		SUBSCRIBE,
		ANNOUNCE,
		TRACK_STATUS_REQUEST,
		TRACK_STATUS,
		GOAWAY,
		SUBSCRIBE_NAMESPACE,
		CLIENT_SETUP,
		SERVER_SETUP:
		messageID = MessageID(num)
	default:
		return 0, nil, errors.New("undefined Message ID")
	}

	// Get a payload reader
	num, err = quicvarint.Read(r)
	if err != nil {
		return 0, nil, err
	}

	reader := io.LimitReader(r, int64(num))

	return messageID, quicvarint.NewReader(reader), nil
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
