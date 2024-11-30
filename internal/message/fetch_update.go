package message

import (
	"io"
	"log"

	"github.com/quic-go/quic-go/quicvarint"
)

type FetchUpdateMessage struct {
	SubscriberPriority SubscriberPriority
}

// TODO
func (fum FetchUpdateMessage) Encode(w io.Writer) error {
	/*
	 * Serialize the message in the following format
	 */

	/*
	 * Serialize the payload
	 */
	p := make([]byte, 0, 1<<4)

	p = quicvarint.Append(p, uint64(fum.SubscriberPriority))

	log.Print("FETCH_UPDATE payload", p) // TODO: delete

	/*
	 * Get serialized message
	 */
	b := make([]byte, 0, len(p)+8)

	// Append the length of the payload
	b = quicvarint.Append(b, uint64(len(p)))

	// Append the payload
	b = append(b, p...)

	// Write
	_, err := w.Write(b)

	return err
}

func (fum *FetchUpdateMessage) Decode(r Reader) error {
	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}

	fum.SubscriberPriority = SubscriberPriority(num)

	return nil
}
