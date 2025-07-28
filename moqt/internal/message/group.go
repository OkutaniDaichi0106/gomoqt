package message

import (
	"bytes"
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

type GroupSequence = protocol.GroupSequence

type GroupMessage struct {
	SubscribeID   SubscribeID
	GroupSequence GroupSequence
}

func (g GroupMessage) Len() int {
	var l int

	l += VarintLen(uint64(g.SubscribeID))
	l += VarintLen(uint64(g.GroupSequence))

	return l
}

func (g GroupMessage) Encode(w io.Writer) error {
	msgLen := g.Len()
	b := pool.Get(msgLen + quicvarint.Len(uint64(msgLen)))
	defer pool.Put(b)

	b = quicvarint.Append(b, uint64(msgLen))
	b = quicvarint.Append(b, uint64(g.SubscribeID))
	b = quicvarint.Append(b, uint64(g.GroupSequence))

	_, err := w.Write(b)

	return err
}

func (g *GroupMessage) Decode(src io.Reader) error {
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
	g.SubscribeID = SubscribeID(num)

	num, err = ReadVarint(r)
	if err != nil {
		return err
	}
	g.GroupSequence = GroupSequence(num)

	return nil
}

func (g GroupMessage) Release() {
	// No resources to release for GroupMessage
}
