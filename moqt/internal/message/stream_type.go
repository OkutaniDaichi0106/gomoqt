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
	slog.Debug("encoding a STREAM_TYPE message")

	/*
	 * Serialize the message in the following format
	 *
	 * STREAM_TYPE Message {
	 *   Stream Type (byte),
	 * }
	 */

	// Write
	_, err := w.Write([]byte{byte(stm.StreamType)})
	if err != nil {
		slog.Error("failed to write a STREAM_TYPE message", slog.String("error", err.Error()))
		return err
	}

	slog.Debug("encoded a STREAM_TYPE message")

	return nil
}

func (stm *StreamTypeMessage) Decode(r io.Reader) error {
	slog.Debug("decoding a STREAM_TYPE message")

	// Get a Stream Type
	buf := make([]byte, 1)
	_, err := r.Read(buf)
	if err != nil {
		slog.Error("failed to read a stream type", slog.String("error", err.Error()))
		return err
	}
	stm.StreamType = StreamType(buf[0])

	slog.Debug("decoded a STREAM_TYPE message")

	return nil
}
