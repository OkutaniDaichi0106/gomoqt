package message

import (
	"io"
	"log/slog"
)

type StreamType byte

/*
 * Serialize the message in the following format
 *
 * STREAM_TYPE Message {
 *   Stream Type (byte),
 * }
 */

type StreamTypeMessage struct {
	StreamType StreamType
}

func (stm StreamTypeMessage) Encode(w io.Writer) (int, error) {

	// Write the Stream Type
	return w.Write([]byte{byte(stm.StreamType)})
}

func (stm *StreamTypeMessage) Decode(r io.Reader) (int, error) {
	// Read the Stream Type
	buf := make([]byte, 1)
	n, err := r.Read(buf)
	if err != nil {
		slog.Error("failed to read a stream type", slog.String("error", err.Error()))
		return n, err
	}
	stm.StreamType = StreamType(buf[0])

	return n, nil
}
