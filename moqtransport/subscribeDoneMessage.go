package moqtransport

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeDoneMessage struct {
	/*
	 * A number to identify the subscribe session
	 */
	subscribeID
	StatusCode    SubscribeDoneStatusCode
	Reason        string
	ContentExists bool
	FinalGroupID  groupID
	FinalObjectID objectID
}

func (sd SubscribeDoneMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * SUBSCRIBE_DONE Message {
	 *   Subscribe ID (varint),
	 *   Status Code (varint),
	 *   Reason ([]byte),
	 *   Content Exist (flag),
	 *   Final Group ID (varint),
	 *   Final Object ID (varint),
	 * }
	 */

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(SUBSCRIBE_DONE))

	// Append Subscirbe ID
	b = quicvarint.Append(b, uint64(sd.subscribeID))

	// Append Status Code
	b = quicvarint.Append(b, uint64(sd.StatusCode))

	// Append Reason
	b = quicvarint.Append(b, uint64(len(sd.Reason)))
	b = append(b, []byte(sd.Reason)...)

	// Append Content Exist
	if !sd.ContentExists {
		b = quicvarint.Append(b, 0)
		return b
	} else if sd.ContentExists {
		b = quicvarint.Append(b, 1)
		// Append the End Group ID only when the Content Exist is true
		b = quicvarint.Append(b, uint64(sd.FinalGroupID))
		// Append the End Object ID only when the Content Exist is true
		b = quicvarint.Append(b, uint64(sd.FinalObjectID))
	}

	return b
}

// func (sd *SubscribeDoneMessage) deserialize(r quicvarint.Reader) error {
// 	// Get Message ID and check it
// 	id, err := deserializeHeader(r)
// 	if err != nil {
// 		return err
// 	}
// 	if id != SUBSCRIBE_DONE {
// 		return errors.New("unexpected message")
// 	}

// 	return sd.deserializeBody(r)
// }

func (sd *SubscribeDoneMessage) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	sd.subscribeID = subscribeID(num)

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	sd.StatusCode = SubscribeDoneStatusCode(num)

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	buf := make([]byte, num)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	sd.Reason = string(buf)

	// Get Content Exist
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	if num == 0 {
		sd.ContentExists = false
		return nil
	} else if num == 1 {
		sd.ContentExists = true
	} else {
		// TODO: terminate the session with a Protocol Violation
		return errors.New("detected value is not flag which takes 0 or 1")
	}

	// Get Largest Group ID only when Content Exist is true
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	sd.FinalGroupID = groupID(num)

	// Get Largest Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	sd.FinalObjectID = objectID(num)

	return nil
}
