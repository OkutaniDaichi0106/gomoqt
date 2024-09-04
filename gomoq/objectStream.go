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
)

// Canonical Object Model
type Object struct {
	/*
	 * Track's namespace
	 */
	TrackNameSpace string

	/*
	 * Track's name
	 */
	TrackName string

	/*
	 * An 8 bit integer indicating the publisher's priority for the Object
	 */
	PublisherPriority

	/*
	 * Forwarding Preference
	 */
	ForwardingPreference

	/*
	 * A number indicating the status of this object
	 * This used to indicate missing objects or mark the end of a group or track
	 */
	StatusCode ObjectStatusCode

	GroupChunk GroupChunk
}

/*
 * OBJECT_STREAM is single object on a unidirectional stream
 * and must be the only message on the unidirectional stream
 */
// type ObjectStream struct {
// 	Object
// 	SubscribeID
// 	TrackAlias
// }

// func (os ObjectStream) serialize() []byte {
// 	/*
// 	 * Serialize as following formatt
// 	 *
// 	 * OBJECT_STREAM Message {
// 	 *   Subscribe ID (varint),
// 	 *   Track Alias (varint),
// 	 *   Group ID (varint),
// 	 *   Object ID (varint),
// 	 *   Publisher Priority (8),
// 	 *   Object Status (varint),
// 	 *   Object Payload (..),
// 	 *}
// 	 */

// 	// TODO?: Chech URI exists

// 	// TODO: Tune the length of the "b"
// 	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
// 	// Append the type of the message
// 	b = quicvarint.Append(b, uint64(OBJECT_STREAM))
// 	// Append Subscribe ID
// 	b = quicvarint.Append(b, uint64(os.SubscribeID))
// 	// Append Track Alias
// 	b = quicvarint.Append(b, uint64(os.TrackAlias))
// 	// Append Group ID
// 	b = quicvarint.Append(b, uint64(os.GroupChunk.GroupID))
// 	// Append Object ID
// 	b = quicvarint.Append(b, uint64(os.GroupChunk.ObjectID))
// 	// Append Publisher Priority
// 	b = quicvarint.Append(b, uint64(os.PublisherPriority))
// 	// Append Object Status Code
// 	b = quicvarint.Append(b, uint64(os.StatusCode))

// 	// Append Object Payload
// 	b = append(b, os.GroupChunk.Payload...)

// 	return b
// }

// func (os *ObjectStream) deserialize(r quicvarint.Reader) error {
// 	// Get Message ID and check it
// 	id, err := deserializeHeader(r)
// 	if err != nil {
// 		return err
// 	}
// 	if id != OBJECT_STREAM {
// 		return errors.New("unexpected message")
// 	}

// 	return os.deserializeBody(r)
// }

// func (os *ObjectStream) deserializeBody(r quicvarint.Reader) error {
// 	var err error
// 	var num uint64

// 	// Stream
// 	os.ForwardingPreference = OBJECT

// 	// Get Subscribe ID
// 	num, err = quicvarint.Read(r)
// 	if err != nil {
// 		return err
// 	}
// 	os.SubscribeID = SubscribeID(num)

// 	// Get Track Alias
// 	num, err = quicvarint.Read(r)
// 	if err != nil {
// 		return err
// 	}
// 	os.TrackAlias = TrackAlias(num)

// 	// TODO?: Get Track Namespace and Track Name from Track Alias

// 	// Get Group ID
// 	num, err = quicvarint.Read(r)
// 	if err != nil {
// 		return err
// 	}
// 	os.GroupChunk.GroupID = GroupID(num)

// 	// Get Object ID
// 	num, err = quicvarint.Read(r)
// 	if err != nil {
// 		return err
// 	}
// 	os.GroupChunk.ObjectID = ObjectID(num)

// 	// Get Publisher Priority
// 	num, err = quicvarint.Read(r)
// 	if err != nil {
// 		return err
// 	}
// 	os.PublisherPriority = PublisherPriority(num)

// 	// Get Object Status Code
// 	num, err = quicvarint.Read(r)
// 	if err != nil {
// 		return err
// 	}
// 	os.StatusCode = ObjectStatusCode(num)

// 	// Get Object Payload
// 	buf, err := io.ReadAll(r)
// 	if err != nil {
// 		return err
// 	}
// 	os.GroupChunk.Payload = buf

// 	return nil
// }
