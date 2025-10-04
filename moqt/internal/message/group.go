package message

import (
	"errors"
	"io"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
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
	b := pool.Get(msgLen + VarintLen(uint64(msgLen)))
	defer pool.Put(b)

	b, _ = WriteVarint(b, uint64(msgLen))
	b, _ = WriteVarint(b, uint64(g.SubscribeID))
	b, _ = WriteVarint(b, uint64(g.GroupSequence))

	_, err := w.Write(b)

	return err
}

func (g *GroupMessage) Decode(src io.Reader) error {
	num, err := ReadMessageLength(src)
	if err != nil {
		return err
	}

	b := pool.Get(int(num))[:num]
	defer pool.Put(b)

	_, err = io.ReadFull(src, b)
	if err != nil {
		return err
	}

	num, n, err := ReadVarint(b)
	if err != nil {
		return err
	}
	g.SubscribeID = SubscribeID(num)
	b = b[n:]

	num, n, err = ReadVarint(b)
	if err != nil {
		return err
	}
	g.GroupSequence = GroupSequence(num)
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}

var ErrMessageTooShort = errors.New("message too short")
