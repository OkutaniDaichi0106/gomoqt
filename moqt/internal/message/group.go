package message

import (
	"bytes"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type GroupSequence uint64

type GroupMessage struct {
	SubscribeID   SubscribeID
	GroupSequence GroupSequence
}

func (g GroupMessage) Len() int {
	l := 0
	l += numberLen(uint64(g.SubscribeID))
	l += numberLen(uint64(g.GroupSequence))
	return l
}

func (g GroupMessage) Encode(w io.Writer) (int, error) {
	p := getBytes()
	defer putBytes(p)

	p = AppendNumber(p, uint64(g.Len()))
	p = AppendNumber(p, uint64(g.SubscribeID))
	p = AppendNumber(p, uint64(g.GroupSequence))

	return w.Write(p)
}

func (g *GroupMessage) Decode(r io.Reader) (int, error) {

	buf, n, err := ReadBytes(quicvarint.NewReader(r))
	if err != nil {
		return n, err
	}

	mr := bytes.NewReader(buf)

	num, _, err := ReadNumber(mr)
	if err != nil {
		return n, err
	}
	g.SubscribeID = SubscribeID(num)

	num, _, err = ReadNumber(mr)
	if err != nil {
		return n, err
	}
	g.GroupSequence = GroupSequence(num)

	return n, nil
}
