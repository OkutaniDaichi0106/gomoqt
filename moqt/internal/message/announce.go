package message

import (
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

func (am AnnounceMessage) Encode(w io.Writer) error {

	p := getBytes()
	defer putBytes(p)

	p = AppendNumber(p, uint64(am.AnnounceStatus))
	p = AppendString(p, am.TrackSuffix)

	_, err := w.Write(p)
	return err
}

func (am *AnnounceMessage) Decode(r io.Reader) error {
	status, _, err := ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}
	am.AnnounceStatus = AnnounceStatus(status)

	am.TrackSuffix, _, err = ReadString(quicvarint.NewReader(r))
	if err != nil {
		return err
	}

	return nil
}
