package gomoq

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
