package moqtmessage

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeDoneMessage struct {
	/*
	 * A number to identify the subscribe session
	 */
	SubscribeID
	StatusCode    SubscribeDoneStatusCode
	Reason        string
	ContentExists bool
	FinalGroupID  GroupID
	FinalObjectID ObjectID
}

func (sd SubscribeDoneMessage) Serialize() []byte {
	/*
	 * Serialize the message in the following formatt
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

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<8)

	// Append the Subscirbe ID
	p = quicvarint.Append(p, uint64(sd.SubscribeID))

	// Append the Status Code
	p = quicvarint.Append(p, uint64(sd.StatusCode))

	// Append the Reason Phrase
	p = quicvarint.Append(p, uint64(len(sd.Reason)))
	p = append(p, []byte(sd.Reason)...)

	// Append the Content Exist
	if !sd.ContentExists {
		p = quicvarint.Append(p, 0)
		return p
	} else if sd.ContentExists {
		p = quicvarint.Append(p, 1)
		// Append the End Group ID only when the Content Exist is true
		p = quicvarint.Append(p, uint64(sd.FinalGroupID))
		// Append the End Object ID only when the Content Exist is true
		p = quicvarint.Append(p, uint64(sd.FinalObjectID))
	}

	/*
	 * Serialize the whole message
	 */
	b := make([]byte, 0, len(p)+1<<4)

	// Append the message type
	b = quicvarint.Append(b, uint64(SUBSCRIBE_DONE))

	// Appen the payload
	b = quicvarint.Append(b, uint64(len(p)))
	b = append(b, p...)

	return b
}

func (sd *SubscribeDoneMessage) DeserializePayload(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	sd.SubscribeID = SubscribeID(num)

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
	sd.FinalGroupID = GroupID(num)

	// Get Largest Object ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	sd.FinalObjectID = ObjectID(num)

	return nil
}
