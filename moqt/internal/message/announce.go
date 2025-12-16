package message

import (
	"io"
)

const (
	ENDED  AnnounceStatus = 0x0
	ACTIVE AnnounceStatus = 0x1
	// LIVE   AnnounceStatus = 0x2
)

type AnnounceStatus byte

type AnnounceMessage struct {
	AnnounceStatus AnnounceStatus
	TrackSuffix    string
}

func (am AnnounceMessage) Len() int {
	var l int

	l += VarintLen(uint64(am.AnnounceStatus))
	l += StringLen(am.TrackSuffix)

	return l
}

func (am AnnounceMessage) Encode(w io.Writer) error {
	msgLen := am.Len()

	b := pool.Get(msgLen + VarintLen(uint64(msgLen)))
	defer pool.Put(b)

	b, _ = WriteMessageLength(b, uint64(msgLen))
	b, _ = WriteVarint(b, uint64(am.AnnounceStatus))
	b, _ = WriteString(b, am.TrackSuffix)

	_, err := w.Write(b)

	return err
}

func (am *AnnounceMessage) Decode(src io.Reader) error {
	size, err := ReadMessageLength(src)
	if err != nil {
		return err
	}

	b := pool.Get(int(size))[:size]
	defer pool.Put(b)

	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	num, n, err := ReadVarint(b)
	if err != nil {
		return err
	}
	am.AnnounceStatus = AnnounceStatus(num)
	b = b[n:]

	str, n, err := ReadString(b)
	if err != nil {
		return err
	}
	am.TrackSuffix = str
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}
