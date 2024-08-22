package gomoq

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
)

type UnsubscribeMessage struct {
	/*
	 * A number to identify the subscribe session
	 */
	SubscribeID
}

func (us UnsubscribeMessage) serialize() []byte {
	/*
	 * Serialize as following formatt
	 *
	 * UNSUBSCRIBE Message {
	 *   Subscribe ID (varint),
	 * }
	 */

	// TODO?: Chech URI exists

	// TODO: Tune the length of the "b"
	b := make([]byte, 0, 1<<10) /* Byte slice storing whole data */
	// Append the type of the message
	b = quicvarint.Append(b, uint64(UNSUBSCRIBE))

	// Append Subscirbe ID
	b = quicvarint.Append(b, uint64(us.SubscribeID))

	return b
}

func (us *UnsubscribeMessage) deserialize(r quicvarint.Reader) error {
	// Get Message ID and check it
	id, err := deserializeHeader(r)
	if err != nil {
		return err
	}
	if id != UNSUBSCRIBE {
		return errors.New("unexpected message")
	}

	return us.deserializeBody(r)
}

func (us *UnsubscribeMessage) deserializeBody(r quicvarint.Reader) error {
	var err error
	var num uint64

	// Get Subscribe ID
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}
	us.SubscribeID = SubscribeID(num)

	return nil
}
