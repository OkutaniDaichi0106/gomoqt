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

func (a AnnounceMessage) Len() int {
	l := 0
	l += numberLen(uint64(a.AnnounceStatus))
	l += stringLen(a.TrackSuffix)

	return l
}

func (am AnnounceMessage) Encode(w io.Writer) (int, error) {

	p := GetBytes()
	defer PutBytes(p)

	p = AppendNumber(p, uint64(am.Len()))
	p = AppendNumber(p, uint64(am.AnnounceStatus))
	p = AppendString(p, am.TrackSuffix)

	return w.Write(p)
}

func (am *AnnounceMessage) Decode(r io.Reader) (int, error) {
	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	mr := bytes.NewReader(buf)

	status, _, err := ReadNumber(mr)
	if err != nil {
		return n, err
	}
	am.AnnounceStatus = AnnounceStatus(status)

	am.TrackSuffix, _, err = ReadString(mr)
	if err != nil {
		return n, err
	}

	return n, nil
}
