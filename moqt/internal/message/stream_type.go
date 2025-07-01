package message

import (
	"io"
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

func (stm StreamTypeMessage) Encode(w io.Writer) error {
	// Write the Stream Type
	_, err := w.Write([]byte{byte(stm.StreamType)})
	return err
}

func (stm *StreamTypeMessage) Decode(r io.Reader) error {
	// Read the Stream Type
	buf := make([]byte, 1)
	_, err := r.Read(buf)
	if err != nil {
		return err
	}
	stm.StreamType = StreamType(buf[0])

	return nil
}
