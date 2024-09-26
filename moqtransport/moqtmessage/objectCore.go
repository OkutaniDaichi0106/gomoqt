package moqtmessage

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
 *
 */
type StreamTypeID uint32

const (
	DATAGRAM_ID StreamTypeID = 0x01
	TRACK_ID    StreamTypeID = 0x02
	PEEP_ID     StreamTypeID = 0x04
)

/*
 * Forwarding Preference
 * Following type are defined in the official document
 * TRACK, PEEP, DATAGRAM
 */
type ObjectForwardingPreference interface {
	ID() StreamTypeID
	StreamType() string
}

var _ ObjectForwardingPreference = (*datagram)(nil)

type datagram struct{}

func (datagram) ID() StreamTypeID {
	return DATAGRAM_ID
}
func (datagram) StreamType() string {
	return "DATAGRAM"
}

var _ ObjectForwardingPreference = (*track)(nil)

type track struct{}

func (track) ID() StreamTypeID {
	return TRACK_ID
}

func (track) StreamType() string {
	return "TRACK"
}

var _ ObjectForwardingPreference = (*peep)(nil)

type peep struct{}

func (peep) ID() StreamTypeID {
	return PEEP_ID
}
func (peep) StreamType() string {
	return "PEEP"
}

var (
	DATAGRAM datagram
	TRACK    track
	PEEP     peep
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
