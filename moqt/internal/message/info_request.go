package message

import (
	"io"
	"log"
	"log/slog"
)

type InfoRequestMessage struct {
	/*
	 * Track name
	 */
	TrackPath []string
}

func (irm InfoRequestMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a INFO_REQUEST message")

	/*
	 * Serialize the payload in the following format
	 *
	 * TRACK_STATUS_REQUEST Message Payload {
	 *   Track Namespace (tuple),
	 *   Track Name ([]byte),
	 * }
	 */

	p := make([]byte, 0, 1<<8)

	// Append the Track Path
	p = appendStringArray(p, irm.TrackPath)

	log.Print("INFO_REQUEST payload", p)

	// Serialize the whole message
	b := make([]byte, 0, len(p)+8)

	// Append the payload
	b = appendBytes(b, p)

	// Write
	_, err := w.Write(b)
	if err != nil {
		slog.Error("failed to write a INFO_REQUEST message")
		return err
	}

	slog.Debug("encoded a INFO_REQUEST message")

	return nil
}

func (irm *InfoRequestMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a INFO_REQUEST message")

	// Get a message reader
	mr, err := newReader(r)
	if err != nil {
		return err
	}

	// Get the Track Path
	irm.TrackPath, err = readStringArray(mr)
	if err != nil {
		return err
	}

	slog.Debug("decoded a INFO_REQUEST message")

	return nil
}
