package message

import (
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

type GroupSequence = protocol.GroupSequence

type GroupMessage struct {
	SubscribeID   SubscribeID
	GroupSequence GroupSequence
}

func (g GroupMessage) Encode(w io.Writer) error {
	p := getBytes()
	defer putBytes(p)

	p = AppendNumber(p, uint64(g.SubscribeID))
	p = AppendNumber(p, uint64(g.GroupSequence))

	_, err := w.Write(p)
	return err
}

func (g *GroupMessage) Decode(r io.Reader) error {
	num, _, err := ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}
	g.SubscribeID = SubscribeID(num)

	num, _, err = ReadNumber(quicvarint.NewReader(r))
	if err != nil {
		return err
	}
	g.GroupSequence = GroupSequence(num)

	return nil
}
