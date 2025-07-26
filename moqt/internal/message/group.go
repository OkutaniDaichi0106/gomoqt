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

	b = quicvarint.Append(b, uint64(msgLen))
	b = quicvarint.Append(b, uint64(g.SubscribeID))
	b = quicvarint.Append(b, uint64(g.GroupSequence))

	_, err := w.Write(b)

	pool.Put(b)

	return err
}

func (g *GroupMessage) Decode(src io.Reader) error {
	num, err := ReadVarint(src)
	if err != nil {
		return err
	}

	b := pool.Get(int(num))[:num]
	_, err = io.ReadFull(src, b)
	if err != nil {
		pool.Put(b)
		return err
	}

	r := bytes.NewReader(b)

	num, err = ReadVarint(r)
	if err != nil {
		pool.Put(b)
		return err
	}
	g.SubscribeID = SubscribeID(num)

	num, err = ReadVarint(r)
	if err != nil {
		pool.Put(b)
		return err
	}
	g.GroupSequence = GroupSequence(num)

	pool.Put(b)

	return nil
}

func (g GroupMessage) Release() {
	// No resources to release for GroupMessage
}
