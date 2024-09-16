package moqtmessage

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type StreamType uint

const (
	//OBJECT_STREAM       MessageID = 0x00 // Deprecated
	OBJECT_DATAGRAM     StreamType = 0x01
	STREAM_HEADER_TRACK StreamType = 0x50
	STREAM_HEADER_GROUP StreamType = 0x51
	STREAM_HEADER_PEEP  StreamType = 0x52
)

/*
 * Deserialize the header of the message which is message id
 */
func DeserializeStreamType(r quicvarint.Reader) (StreamType, error) {
	// Get the first number in the message expected to be MessageID
	num, err := quicvarint.Read(r)
	if err != nil {
		return 0xff, err
	}
	switch StreamType(num) {
	case
		STREAM_HEADER_TRACK,
		STREAM_HEADER_PEEP:
		return StreamType(num), nil
	default:
		return 0xff, errors.New("undefined Message ID")
	}
}

// Group ID
type GroupID uint64

// Peep ID
type PeepID uint64

// Object ID
type ObjectID uint64

/*
 * Track Alias
 * This must be more than 1
 * If This is 0, throw an error
 */
type TrackAlias uint64

// Publisher Priority
type PublisherPriority byte

/*
 * Forwarding Preference
 * Following type are defined in the official document
 * TRACK, GROUP, OBJECT, DATAGRAM
 */
type ForwardingPreference uint64

const (
	TRACK ForwardingPreference = iota
	//GROUP
	//OBJECT
	PEEP
	DATAGRAM
)

// Object Status
type ObjectStatusCode uint64

const (
	NOMAL_OBJECT       ObjectStatusCode = 0x00
	NONEXISTENT_OBJECT ObjectStatusCode = 0x01
	NONEXISTENT_GROUP  ObjectStatusCode = 0x02
	END_OF_GROUP       ObjectStatusCode = 0x03
	END_OF_TRACK       ObjectStatusCode = 0x04
	END_OF_PEEP        ObjectStatusCode = 0x05
)
