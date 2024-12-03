package message

import (
	"io"
	"log/slog"
)

type StreamType byte

type StreamTypeMessage struct {
	StreamType StreamType
}

func (stm StreamTypeMessage) Encode(w io.Writer) error {
	slog.Debug("encoding a SESSION_UPDATE message")

	/*
	 * Serialize the message in the following format
	 *
	 * STREAM_TYPE Message {
	 *   Stream Type (byte),
	 * }
	 */

	// Write
	_, err := w.Write([]byte{byte(stm.StreamType)})

	return err
}

func (stm *StreamTypeMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a SESSION_UPDATE message")

	// Get a Stream Type
	buf := make([]byte, 1)
	_, err := r.Read(buf)
	if err != nil {
		return err
	}
	stm.StreamType = StreamType(buf[0])

	return nil
}
