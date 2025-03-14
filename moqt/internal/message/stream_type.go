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

func (stm StreamTypeMessage) Len() int {
	return 1
}

func (stm StreamTypeMessage) Encode(w io.Writer) (int, error) {
	slog.Debug("encoding a STREAM_TYPE message")
	// Write the Stream Type
	n, err := w.Write([]byte{byte(stm.StreamType)})
	if err != nil {
		slog.Error("failed to write a stream type", slog.String("error", err.Error()))
		return n, err
	}

	slog.Debug("encoded a STREAM_TYPE message")

	return n, nil
}

func (stm *StreamTypeMessage) Decode(r io.Reader) (int, error) {
	slog.Debug("decoding a STREAM_TYPE message")

	// Read the Stream Type
	buf := make([]byte, 1)
	n, err := r.Read(buf)
	if err != nil {
		slog.Error("failed to read a stream type", slog.String("error", err.Error()))
		return n, err
	}
	stm.StreamType = StreamType(buf[0])

	slog.Debug("decoded a STREAM_TYPE message")

	return n, nil
}
