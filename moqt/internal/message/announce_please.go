package message

import (
	"io"
)

/*
 *	ANNOUNCE_PLEASE Message {
 *	  Track Prefix (string),
 *	}
 */
type AnnouncePleaseMessage struct {
	TrackPrefix string
}

func (aim AnnouncePleaseMessage) Len() int {
	return StringLen(aim.TrackPrefix)
}

func (aim AnnouncePleaseMessage) Encode(w io.Writer) error {
	msgLen := aim.Len()
	b := make([]byte, 0, msgLen+VarintLen(uint64(msgLen)))

	b, _ = WriteMessageLength(b, uint64(msgLen))
	b, _ = WriteString(b, aim.TrackPrefix)

	_, err := w.Write(b)

	return err
}

func (aim *AnnouncePleaseMessage) Decode(src io.Reader) error {
	num, err := ReadMessageLength(src)
	if err != nil {
		return err
	}

	b := make([]byte, num)

	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	str, n, err := ReadString(b)
	if err != nil {
		return err
	}
	aim.TrackPrefix = str
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}
