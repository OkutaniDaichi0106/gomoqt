package message

import (
	"io"
	"log"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type FetchUpdateMessage struct {
	SubscriberPriority SubscriberPriority
}

func (fum FetchUpdateMessage) Encode(w io.Writer) error {
	slog.Debug("decoding a FETCH_UPDATE message")

	/*
	 * Serialize the message in the following format
	 *
	 * FETCH_UPDATE Message Payload {
	 *   Subscriber Priority (varint),
	 * }
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
	if err != nil {
		return err
	}

	slog.Debug("encoded a FETCH_UPDATE message")

	return nil
}

func (fum *FetchUpdateMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a FETCH_UPDATE message")

	// Get a messaga reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	num, err := quicvarint.Read(mr)
	if err != nil {
		return err
	}

	fum.SubscriberPriority = SubscriberPriority(num)

	slog.Debug("decoded a FETCH_UPDATE message")

	return nil
}
