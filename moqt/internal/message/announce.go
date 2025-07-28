package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
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

	b := pool.Get(msgLen)
	defer pool.Put(b)

	b = quicvarint.Append(b, uint64(msgLen))
	b = quicvarint.Append(b, uint64(am.AnnounceStatus))
	b = quicvarint.Append(b, uint64(len(am.TrackSuffix)))
	b = append(b, am.TrackSuffix...)

	_, err := w.Write(b)

	return err
}

func (am *AnnounceMessage) Decode(src io.Reader) error {
	num, err := ReadVarint(src)
	if err != nil {
		return err
	}

	b := pool.Get(int(num))[:num]
	defer pool.Put(b)

	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	r := bytes.NewReader(b)

	num, err = ReadVarint(r)
	if err != nil {
		return err
	}
	am.AnnounceStatus = AnnounceStatus(num)

	str, err := ReadString(r)
	if err != nil {
		return err
	}
	am.TrackSuffix = str

	return nil
}
