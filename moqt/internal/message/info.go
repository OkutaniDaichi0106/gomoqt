package message

import (
	"bytes"
	"io"
	"log/slog"

	"github.com/quic-go/quic-go/quicvarint"
)

type InfoMessage struct {
	TrackPriority       TrackPriority
	LatestGroupSequence GroupSequence
	GroupOrder          GroupOrder
}

func (im InfoMessage) Len() int {
	l := 0
	l += quicvarint.Len(uint64(im.TrackPriority))
	l += quicvarint.Len(uint64(im.LatestGroupSequence))
	l += quicvarint.Len(uint64(im.GroupOrder))
	return l
}

func (im InfoMessage) Encode(w io.Writer) (int, error) {

	p := GetBytes()
	defer PutBytes(p)

	p = AppendNumber(p, uint64(im.Len()))
	p = AppendNumber(p, uint64(im.TrackPriority))
	p = AppendNumber(p, uint64(im.LatestGroupSequence))
	p = AppendNumber(p, uint64(im.GroupOrder))

	n, err := w.Write(p)
	if err != nil {
		slog.Error("failed to write a INFO message", "error", err)
		return n, err
	}

	slog.Debug("encoded a INFO message")

	return n, err
}

func (im *InfoMessage) Decode(r io.Reader) (int, error) {

	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := ReadNumber(mr)
	if err != nil {
		return n, err
	}
	im.TrackPriority = TrackPriority(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		return n, err
	}
	im.LatestGroupSequence = GroupSequence(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		return n, err
	}
	im.GroupOrder = GroupOrder(num)

	slog.Debug("decoded a INFO message")

	return n, nil
}
