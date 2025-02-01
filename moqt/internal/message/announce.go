package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	ENDED  AnnounceStatus = 0x0
	ACTIVE AnnounceStatus = 0x1
	LIVE   AnnounceStatus = 0x2
)

type AnnounceStatus byte

type AnnounceMessage struct {
	AnnounceStatus     AnnounceStatus
	TrackSuffix        []string
	AnnounceParameters Parameters
}

func (a AnnounceMessage) Len() int {
	l := 0
	l += numberLen(uint64(a.AnnounceStatus))
	l += stringArrayLen(a.TrackSuffix)
	l += parametersLen(a.AnnounceParameters)
	return l
}

func (a AnnounceMessage) Encode(w io.Writer) (int, error) {
	p := GetBytes()
	defer PutBytes(p)

	*p = AppendNumber(*p, uint64(a.Len()))
	*p = AppendNumber(*p, uint64(a.AnnounceStatus))
	*p = AppendStringArray(*p, a.TrackSuffix)
	*p = AppendParameters(*p, a.AnnounceParameters)

	return w.Write(*p)
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

	am.TrackSuffix, _, err = ReadStringArray(mr)
	if err != nil {
		return n, err
	}

	am.AnnounceParameters, _, err = ReadParameters(mr)
	if err != nil {
		return n, err
	}

	return n, nil
}
