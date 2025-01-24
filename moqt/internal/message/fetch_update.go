package message

import (
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type FetchUpdateMessage struct {
	TrackPriority TrackPriority
}

func (fum FetchUpdateMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a FETCH_UPDATE message")

	// Serialize the payload
	payload := quicvarint.Append(nil, uint64(fum.TrackPriority))

	// Serialize the message with the length of the payload
	message := quicvarint.Append(nil, uint64(len(payload)))
	message = append(message, payload...)

	// Write the serialized message
	_, err := w.Write(message)
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

	fum.TrackPriority = TrackPriority(num)

	slog.Debug("decoded a FETCH_UPDATE message")

	return nil
}
