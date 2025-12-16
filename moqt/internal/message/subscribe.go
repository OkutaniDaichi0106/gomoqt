package message

import (
	"io"
)

/*
* SUBSCRIBE Message {
*   Subscribe ID (varint),
*   Broadcast Path (string),
*   Track Name (string),
*   Track Priority (varint),
* }
 */
type SubscribeMessage struct {
	SubscribeID   uint64
	BroadcastPath string
	TrackName     string
	TrackPriority uint8
}

func (s SubscribeMessage) Len() int {
	var l int

	l += VarintLen(uint64(s.SubscribeID))
	l += StringLen(s.BroadcastPath)
	l += StringLen(s.TrackName)
	l += VarintLen(uint64(s.TrackPriority))

	return l
}

func (s SubscribeMessage) Encode(w io.Writer) error {
	msgLen := s.Len()
	b := make([]byte, 0, msgLen+VarintLen(uint64(msgLen)))

	b, _ = WriteMessageLength(b, uint64(msgLen))
	b, _ = WriteVarint(b, uint64(s.SubscribeID))
	b, _ = WriteVarint(b, uint64(len(s.BroadcastPath)))
	b = append(b, s.BroadcastPath...)
	b, _ = WriteVarint(b, uint64(len(s.TrackName)))
	b = append(b, s.TrackName...)
	b, _ = WriteVarint(b, uint64(s.TrackPriority))

	_, err := w.Write(b)
	return err
}

func (s *SubscribeMessage) Decode(src io.Reader) error {
	size, err := ReadMessageLength(src)
	if err != nil {
		return err
	}

	b := make([]byte, size)

	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	num, n, err := ReadVarint(b)
	if err != nil {
		return err
	}
	s.SubscribeID = num
	b = b[n:]

	str, n, err := ReadString(b)
	if err != nil {
		return err
	}
	s.BroadcastPath = str
	b = b[n:]

	str, n, err = ReadString(b)
	if err != nil {
		return err
	}
	s.TrackName = str
	b = b[n:]

	num, n, err = ReadVarint(b)
	if err != nil {
		return err
	}
	s.TrackPriority = uint8(num)
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}
